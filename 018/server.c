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

struct bpf_t
{
    int interface_index;
    struct xdp_program * program;
    bool attached_native;
    bool attached_skb;
    int counters_fd;
    int server_stats_fd;
};

static struct bpf_t bpf;

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

    // load the server_xdp program and attach it to the network interface

    printf( "loading server_xdp...\n" );

    bpf->program = xdp_program__open_file( "server_xdp.o", "xdp", NULL );
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

    // fork workers

    for ( int i = 0; i < MAX_CPUS; i++ )
    {   
        pid_t c = fork();
        if ( c == 0 )
        { 
            // child worker process
            printf( "starting golang worker %d\n", i );
            fflush( stdout );
            char cpu_string[64];
            sprintf( cpu_string, "%d", i );
            char * args[] = { "taskset", "-c", cpu_string, "./worker", cpu_string, 0 };
            execv( "/usr/bin/taskset", args );
            exit(0); 
        } 
    }

    // main loop

    pin_thread_to_cpu( MAX_CPUS );       // IMPORTANT: keep out of the way of the XDP cpus on google cloud [0,15]

    unsigned int num_cpus = libbpf_num_possible_cpus();

    uint64_t previous_player_state_packets_sent = 0;

    while ( !quit )
    {
        usleep( 1000000 );

        // track stats

        struct counters values[num_cpus];
        int key = 0;
        if ( bpf_map_lookup_elem( bpf.counters_fd, &key, values ) != 0 ) 
        {
            printf( "\nerror: could not look up counters: %s\n\n", strerror( errno ) );
            quit = true;
            break;
        }

        uint64_t current_player_state_packets_sent = 0;

        for ( int i = 0; i < MAX_CPUS; i++ )
        {
            current_player_state_packets_sent += values[i].player_state_packets_sent;
        }

        // print out important stats

        uint64_t player_state_delta = current_player_state_packets_sent - previous_player_state_packets_sent;

        printf( "player state delta: %" PRId64 "\n", player_state_delta );

        previous_player_state_packets_sent = current_player_state_packets_sent;

        // upload stats to the xdp program to be sent down to clients

        struct server_stats stats;
        memset( &stats, 0, sizeof(stats) );
        stats.player_state_packets_sent = current_player_state_packets_sent;

        int err = bpf_map_update_elem( bpf.server_stats_fd, &key, &stats, BPF_ANY );
        if ( err != 0 )
        {
            printf( "\nerror: failed to update server stats: %s\n\n", strerror(errno) );
            quit = true;
            break;
        }
    }

    cleanup();

    return 0;
}
