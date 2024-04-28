/*
    FPS server XDP program (Userspace)

    Runs on Ubuntu 22.04 LTS 64bit with Linux Kernel 6.5+ *ONLY*
*/

#define _GNU_SOURCE

#include <memory.h>
#include <stdio.h>
#include <signal.h>
#include <errno.h>
#include <stdbool.h>
#include <assert.h>
#include <unistd.h>
#include <ifaddrs.h>
#include <net/if.h>
#include <bpf/bpf.h>
#include <bpf/libbpf.h>
#include <xdp/libxdp.h>
#include <sys/resource.h>
#include <sys/types.h>
#include <inttypes.h>
#include <time.h>
#include <pthread.h>
#include <sched.h>
#include <stdlib.h>
#include "shared.h"
#include "map.h"

struct bpf_t
{
    int interface_index;
    struct xdp_program * program;
    bool attached_native;
    bool attached_skb;
    int counters_fd;
    int server_stats_fd;
    int input_buffer_outer_fd;
    int input_buffer_inner_fd[MAX_CPUS];
    int player_state_outer_fd;
    int player_state_inner_fd[MAX_CPUS];
    struct ring_buffer * input_buffer[MAX_CPUS];
    int ring_buffer_cpus[MAX_CPUS];
};

static struct bpf_t bpf;

static uint64_t inputs_processed[MAX_CPUS];
static uint64_t inputs_lost[MAX_CPUS];

static struct map_t * cpu_player_map[MAX_CPUS];

static int process_input( void * ctx, void * data, size_t data_sz )
{
    int cpu = *(int*) ctx;

    // todo
    printf( "process input on cpu %d\n", cpu );

    struct input_header * header = (struct input_header*) data;

    struct input_data * input = (struct input_data*) data + sizeof(struct input_header);

    struct player_state * state = map_get( cpu_player_map[cpu], header->session_id );
    if ( !state )
    {
        // first player update
        state = malloc( sizeof(struct player_state) );
        map_set( cpu_player_map[cpu], header->session_id, state );
    }

    // todo: handle multiple inputs

    state->t += input->dt;

    for ( int i = 0; i < PLAYER_STATE_SIZE; i++ )
    {
        state->data[i] = (uint8_t) state->t + (uint8_t) i;
    }

    int player_state_fd = bpf.player_state_inner_fd[cpu];

    int err = bpf_map_update_elem( player_state_fd, &header->session_id, state, BPF_ANY );
    if ( err != 0 )
    {
        printf( "error: failed to update player state: %s\n", strerror(errno) );
        return 0;
    }

    __sync_fetch_and_add( &inputs_processed[cpu], 1 );

    return 0;
}

static double time_start;

void platform_init()
{
    struct timespec ts;
    clock_gettime( CLOCK_MONOTONIC_RAW, &ts );
    time_start = ts.tv_sec + ( (double) ( ts.tv_nsec ) ) / 1000000000.0;
}

double platform_time()
{
    struct timespec ts;
    clock_gettime( CLOCK_MONOTONIC_RAW, &ts );
    double current = ts.tv_sec + ( (double) ( ts.tv_nsec ) ) / 1000000000.0;
    return current - time_start;
}

void platform_sleep( double time )
{
    usleep( (int) ( time * 1000000 ) );
}

int bpf_init( struct bpf_t * bpf, const char * interface_name )
{
    // we can only run xdp programs as root

    if ( geteuid() != 0 ) 
    {
        printf( "\nerror: this program must be run as root\n\n" );
        return 1;
    }

    // find the network interface that matches the interface name
    {
        bool found = false;

        struct ifaddrs * addrs;
        if ( getifaddrs( &addrs ) != 0 )
        {
            printf( "\nerror: getifaddrs failed\n\n" );
            return 1;
        }

        for ( struct ifaddrs * iap = addrs; iap != NULL; iap = iap->ifa_next ) 
        {
            if ( iap->ifa_addr && ( iap->ifa_flags & IFF_UP ) && iap->ifa_addr->sa_family == AF_INET )
            {
                struct sockaddr_in * sa = (struct sockaddr_in*) iap->ifa_addr;
                if ( strcmp( interface_name, iap->ifa_name ) == 0 )
                {
                    printf( "found network interface: '%s'\n", iap->ifa_name );
                    bpf->interface_index = if_nametoindex( iap->ifa_name );
                    if ( !bpf->interface_index ) 
                    {
                        printf( "\nerror: if_nametoindex failed\n\n" );
                        return 1;
                    }
                    found = true;
                    break;
                }
            }
        }

        freeifaddrs( addrs );

        if ( !found )
        {
            printf( "\nerror: could not find any network interface matching '%s'\n\n", interface_name );
            return 1;
        }
    }

    // initialize platform

    platform_init();

    // load the server_xdp program and attach it to the network interface

    printf( "loading server_xdp...\n" );

    bpf->program = xdp_program__open_file( "server_xdp.o", "server_xdp", NULL );
    if ( libxdp_get_error( bpf->program ) ) 
    {
        printf( "\nerror: could not load server_xdp program\n\n");
        return 1;
    }

    printf( "server_xdp loaded successfully.\n" );

    printf( "attaching server_xdp to network interface\n" );

    int ret = xdp_program__attach( bpf->program, bpf->interface_index, XDP_MODE_NATIVE, 0 );
    if ( ret == 0 )
    {
        bpf->attached_native = true;
    } 
    else
    {
        printf( "falling back to skb mode...\n" );
        ret = xdp_program__attach( bpf->program, bpf->interface_index, XDP_MODE_SKB, 0 );
        if ( ret == 0 )
        {
            bpf->attached_skb = true;
        }
        else
        {
            printf( "\nerror: failed to attach server_xdp program to interface\n\n" );
            return 1;
        }
    }

    // bump rlimit

    struct rlimit rlim_new = {
        .rlim_cur   = RLIM_INFINITY,
        .rlim_max   = RLIM_INFINITY,
    };

    if ( setrlimit( RLIMIT_MEMLOCK, &rlim_new ) ) 
    {
        printf( "\nerror: could not increase RLIMIT_MEMLOCK limit!\n\n" );
        return 1;
    }

    // get the file handle to counters

    bpf->counters_fd = bpf_obj_get( "/sys/fs/bpf/counters_map" );
    if ( bpf->counters_fd <= 0 )
    {
        printf( "\nerror: could not get counters: %s\n\n", strerror(errno) );
        return 1;
    }

    // get the file handle to the server stats

    bpf->server_stats_fd = bpf_obj_get( "/sys/fs/bpf/server_stats" );
    if ( bpf->server_stats_fd <= 0 )
    {
        printf( "\nerror: could not get server stats: %s\n\n", strerror(errno) );
        return 1;
    }

    // get the file handle to the outer player state map

    bpf->player_state_outer_fd = bpf_obj_get( "/sys/fs/bpf/player_state_map" );
    if ( bpf->player_state_outer_fd <= 0 )
    {
        printf( "\nerror: could not get outer player state map: %s\n\n", strerror(errno) );
        return 1;
    }

    // get the file handle to the inner player state maps

    for ( int i = 0; i < MAX_CPUS; i++ )
    {
        uint32_t key = i;
        uint32_t inner_map_id = 0;
        int result = bpf_map_lookup_elem( bpf->player_state_outer_fd, &key, &inner_map_id );
        if ( result != 0 )
        {
            printf( "\nerror: failed lookup player state inner map: %s\n\n", strerror(errno) );
            return 1;
        }
        bpf->player_state_inner_fd[i] = bpf_map_get_fd_by_id( inner_map_id );
        printf( "player state for cpu %d = %d\n", i, bpf->player_state_inner_fd[i] );
    }

    // get the file handle to the outer input buffer map

    bpf->input_buffer_outer_fd = bpf_obj_get( "/sys/fs/bpf/input_buffer_map" );
    if ( bpf->input_buffer_outer_fd <= 0 )
    {
        printf( "\nerror: could not get outer input buffer map: %s\n\n", strerror(errno) );
        return 1;
    }

    // get the file handle to the inner input buffer maps

    for ( int i = 0; i < MAX_CPUS; i++ )
    {
        uint32_t key = i;
        uint32_t inner_map_id = 0;
        int result = bpf_map_lookup_elem( bpf->input_buffer_outer_fd, &key, &inner_map_id );
        if ( result != 0 )
        {
            printf( "\nerror: failed lookup input buffer inner map: %s\n\n", strerror(errno) );
            return 1;
        }
        bpf->input_buffer_inner_fd[i] = bpf_map_get_fd_by_id( inner_map_id );
        printf( "input buffer for cpu %d = %d\n", i, bpf->input_buffer_inner_fd[i] );
    }

    // create the input ring buffer

    for ( int i = 0; i < MAX_CPUS; i++ )
    {
        bpf->ring_buffer_cpus[i] = i;
        bpf->input_buffer[i] = ring_buffer__new( bpf->input_buffer_inner_fd[i], process_input, bpf->ring_buffer_cpus + i, NULL );
        if ( !bpf->input_buffer[i] )
        {
            printf( "\nerror: could not create input buffer[%d]\n\n", i );
            return 1;
        }
    }

    printf( "ready\n" );

    return 0;
}

void bpf_shutdown( struct bpf_t * bpf )
{
    assert( bpf );

    for ( int i = 0; i < MAX_CPUS; i++ )
    {
        if ( bpf->input_buffer[i] )
        {
            ring_buffer__free( bpf->input_buffer[i] );
            bpf->input_buffer[i] = NULL;
        }
    }

    if ( bpf->program != NULL )
    {
        if ( bpf->attached_native )
        {
            xdp_program__detach( bpf->program, bpf->interface_index, XDP_MODE_NATIVE, 0 );
        }
        if ( bpf->attached_skb )
        {
            xdp_program__detach( bpf->program, bpf->interface_index, XDP_MODE_SKB, 0 );
        }
        xdp_program__close( bpf->program );
    }
}

volatile bool quit;

void interrupt_handler( int signal )
{
    (void) signal; quit = true;
}

void clean_shutdown_handler( int signal )
{
    (void) signal;
    quit = true;
}

static void cleanup()
{
    bpf_shutdown( &bpf );
    fflush( stdout );
}

int pin_thread_to_cpu( int cpu ) 
{
    int num_cpus = sysconf(_SC_NPROCESSORS_ONLN );
    if ( cpu < 0 || cpu >= num_cpus  )
        return EINVAL;

    cpu_set_t cpuset;
    CPU_ZERO( &cpuset );
    CPU_SET( cpu, &cpuset );

    pthread_t current_thread = pthread_self();    

    return pthread_setaffinity_np( current_thread, sizeof(cpu_set_t), &cpuset );
}

void * worker_thread_function( void * context )
{
    int cpu = *( (int*) context );

    printf( "worker thread sees cpu is #%d\n" cpu );

    pin_thread_to_cpu( MAX_CPUS + cpu );   // IMPORTANT: Worker threads run on CPUs [16,31], but *logically* work with maps in the CPU range [0,15]

    while ( !quit )
    {
        // poll ring buffer to drive input processing

        int err = ring_buffer__poll( bpf.input_buffer[cpu], 1 );
        if ( err == -EINTR )
        {
            // ctrl-c
            quit = true;
            break;
        }
        if ( err < 0 ) 
        {
            printf( "\nerror: could not poll input buffer: %d\n\n", err );
            quit = true;
            break;
        }
    }

    return NULL;
}

int main( int argc, char *argv[] )
{
    signal( SIGINT,  interrupt_handler );
    signal( SIGTERM, clean_shutdown_handler );
    signal( SIGHUP,  clean_shutdown_handler );

    if ( argc != 2 )
    {
        printf( "\nusage: server <interface name>\n\n" );
        return 1;
    }

    for ( int i = 0; i < MAX_CPUS; i++ )
    {
        cpu_player_map[i] = map_create();
    }

    const char * interface_name = argv[1];

    if ( bpf_init( &bpf, interface_name ) != 0 )
    {
        cleanup();
        return 1;
    }

    // run worker threads

    int thread_cpu[MAX_CPUS];
    pthread_t thread_id[MAX_CPUS];

    for ( int i = 0; i < MAX_CPUS; i++ )
    {
        printf( "starting worker thread %d\n", i );
        thread_cpu[i] = i;
        pthread_create( &thread_id[i], NULL, worker_thread_function, &thread_cpu ); 
    }

    // main loop

    pin_thread_to_cpu( MAX_CPUS );       // IMPORTANT: keep the main thread out of the way of the XDP cpus on google cloud [0,15]

    unsigned int num_cpus = libbpf_num_possible_cpus();

    uint64_t previous_processed_inputs = 0;
    uint64_t previous_player_state_packets_sent = 0;
    uint64_t previous_lost_inputs = 0;

    while ( !quit )
    {
        platform_sleep( 1.0 );

        // track stats

        struct counters values[num_cpus];
        int key = 0;
        if ( bpf_map_lookup_elem( bpf.counters_fd, &key, values ) != 0 ) 
        {
            printf( "\nerror: could not look up counters: %s\n\n", strerror( errno ) );
            quit = true;
            break;
        }

        uint64_t current_processed_inputs = 0;
        uint64_t current_player_state_packets_sent = 0;
        uint64_t current_lost_inputs = 0;
        for ( int i = 0; i < MAX_CPUS; i++ )
        {
            current_processed_inputs += inputs_processed[i];
            current_player_state_packets_sent += values[i].player_state_packets_sent;
            current_lost_inputs += inputs_lost[i];
        }

        // print out important stats

        uint64_t input_delta = current_processed_inputs - previous_processed_inputs;
        uint64_t player_state_delta = current_player_state_packets_sent - previous_player_state_packets_sent;
        uint64_t lost_delta = current_lost_inputs - previous_lost_inputs;
        printf( "input delta: %" PRId64 ", player state delta: %" PRId64 ", lost delta: %" PRId64 "\n", input_delta, player_state_delta, lost_delta );
        previous_processed_inputs = current_processed_inputs;
        previous_player_state_packets_sent = current_player_state_packets_sent;
        previous_lost_inputs = current_lost_inputs;

        // upload stats to the xdp program to be sent down to clients

        struct server_stats stats;
        stats.inputs_processed = current_processed_inputs;
        stats.player_state_packets_sent = current_player_state_packets_sent;

        int err = bpf_map_update_elem( bpf.server_stats_fd, &key, &stats, BPF_ANY );
        if ( err != 0 )
        {
            printf( "\nerror: failed to update server stats: %s\n\n", strerror(errno) );
            quit = true;
            break;
        }
    }

    // clean up

    for ( int i = 0; i < MAX_CPUS; i++ )
    {
        pthread_join( thread_id[i], NULL );
    }

    cleanup();

    printf( "\n" );

    return 0;
}
