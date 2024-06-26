/*
    FPS server XDP program (Userspace)

    Runs on Ubuntu 22.04 LTS 64bit with Linux Kernel 6.5+ *ONLY*
*/

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

#define MAX_CPUS 1024

static uint64_t inputs_processed[MAX_CPUS];

void process_input( void * ctx, int cpu, void * data, unsigned int data_sz )
{
    (void) ctx;
    (void) data;
    __sync_fetch_and_add( &inputs_processed[cpu], 1 );
}

struct bpf_t
{
    int interface_index;
    struct xdp_program * program;
    bool attached_native;
    bool attached_skb;
    int input_buffer_fd;
    int server_stats_fd;
    struct perf_buffer * input_buffer;
};

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
            printf( "\nerror: could not find any network interface matching '%s'", interface_name );
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

    // bump rlimit for the perf buffer

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

    // get the file handle to the server stats

    bpf->server_stats_fd = bpf_obj_get( "/sys/fs/bpf/server_stats" );
    if ( bpf->server_stats_fd <= 0 )
    {
        printf( "\nerror: could not get server stats: %s\n\n", strerror(errno) );
        return 1;
    }

    // create the input perf buffer

    bpf->input_buffer = perf_buffer__new( bpf->input_buffer_fd, 131072, process_input, NULL, NULL, NULL );
    if ( libbpf_get_error( bpf->input_buffer ) ) 
    {
        printf( "\nerror: could not create input buffer\n\n" );
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

#pragma pack(push, 1)

struct server_stats
{
    __u64 inputs_processed;
};

#pragma pack(pop)

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

    double last_print_time = platform_time();

    uint64_t last_inputs = 0;

    while ( !quit )
    {
        int err = perf_buffer__poll( bpf.input_buffer, 1 );
        if ( err < 0 ) 
        {
            printf( "\nerror: could not poll input buffer: %d\n", err );
            quit = true;
            break;
        }

        double current_time = platform_time();
        if ( last_print_time + 1.0 <= current_time )
        {
            uint64_t current_inputs = 0;
            for ( int i = 0; i < MAX_CPUS; i++ )
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
