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

#include "shared.h"

static uint64_t inputs_processed[XDP_MAX_CPUS];

struct bpf_t
{
    int interface_index;
    struct xdp_program * program;
    bool attached_native;
    bool attached_skb;
    int input_buffer_fd;
    struct ring_buffer * input_buffer;
    int server_stats_fd;
    int player_state_outer_fd;
    int player_state_inner_fd[XDP_MAX_CPUS];
};

static int process_input( void * ctx, void * data, size_t data_sz )
{
    // todo: temporary
    int cpu = 0;

    struct bpf_t * bpf = (struct bpf_t*) ctx;

    int player_state_fd = bpf->player_state_inner_fd[cpu];

    struct input_header * header = (struct input_header*) data;

    struct input_data * input = (struct input_data*) data + sizeof(struct input_header);

    struct player_state state;

    uint64_t value;
    int result = bpf_map_lookup_elem( player_state_fd, &header->session_id, &state );
    if ( result != 0 )
    {
        // first player update
        memset( &state, 0, sizeof(struct player_state) );
    }

    // todo: handle multiple inputs

    state.t += input->dt;

    for ( int i = 0; i < PLAYER_STATE_SIZE; i++ )
    {
        state.data[i] = (uint8_t) state.t + (uint8_t) i;
    }

    int err = bpf_map_update_elem( player_state_fd, &header->session_id, &state, BPF_ANY );
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

    // get the file handle to the input buffer

    bpf->input_buffer_fd = bpf_obj_get( "/sys/fs/bpf/input_buffer" );
    if ( bpf->input_buffer_fd <= 0 )
    {
        printf( "\nerror: could not get input buffer: %s\n\n", strerror(errno) );
        return 1;
    }

    // create the input ring buffer

    bpf->input_buffer = ring_buffer__new( bpf->input_buffer_fd, process_input, NULL, NULL );
    if ( !bpf->input_buffer )
    {
        printf( "\nerror: could not create input buffer\n\n" );
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

    for ( int i = 0; i < XDP_MAX_CPUS; i++ )
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
    }

    printf( "ready\n" );

    return 0;
}

void bpf_shutdown( struct bpf_t * bpf )
{
    assert( bpf );

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

static struct bpf_t bpf;

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

int pin_thread_to_core( int core_id ) 
{
   int num_cores = sysconf(_SC_NPROCESSORS_ONLN );
   if ( core_id < 0 || core_id >= num_cores  )
      return EINVAL;

   cpu_set_t cpuset;
   CPU_ZERO( &cpuset );
   CPU_SET( core_id, &cpuset );

   pthread_t current_thread = pthread_self();    

   return pthread_setaffinity_np( current_thread, sizeof(cpu_set_t), &cpuset );
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

    const char * interface_name = argv[1];

    if ( bpf_init( &bpf, interface_name ) != 0 )
    {
        cleanup();
        return 1;
    }

    // double last_print_time = platform_time();

    // uint64_t last_inputs = 0;

    // pin_thread_to_core( 16 );       // IMPORTANT: keep the main thread out of the way of the XDP cpus on google cloud [0,15]

    while ( !quit )
    {
        int err = ring_buffer__poll( bpf.input_buffer, 1 );
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

        // double current_time = platform_time();

        if ( last_print_time + 1.0 <= current_time )
        {
            uint64_t current_inputs = 0;
            for ( int i = 0; i < XDP_MAX_CPUS; i++ )
            {
                current_inputs += inputs_processed[i];
            }
            uint64_t input_delta = current_inputs - last_inputs;
            printf( "input delta: %" PRId64 "\n", input_delta );
            last_inputs = current_inputs;
            last_print_time = current_time;

            struct server_stats stats;
            stats.inputs_processed = current_inputs;

            __u32 key = 0;
            int err = bpf_map_update_elem( bpf.server_stats_fd, &key, &stats, BPF_ANY );
            if ( err != 0 )
            {
                printf( "\nerror: failed to update server stats: %s\n\n", strerror(errno) );
                quit = true;
                break;
            }
        }
    }

    cleanup();

    printf( "\n" );

    return 0;
}

#endif