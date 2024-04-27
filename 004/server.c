
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

static int process_input( void * ctx, void * data, size_t data_sz )
{
    return 0;
}

int main( int argc, char *argv[] )
{
    // we can only run xdp programs as root

    if ( geteuid() != 0 ) 
    {
        printf( "\nerror: this program must be run as root\n\n" );
        return 1;
    }

    // find the network interface that matches the interface name

    if ( argc != 2 )
    {
        printf( "\nusage: server <interface name>\n\n" );
        return 1;
    }

    const char * interface_name = argv[1];

    int interface_index = 0;
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
                    interface_index = if_nametoindex( iap->ifa_name );
                    if ( !interface_index ) 
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

    struct xdp_program * program = NULL;
    
    if ( geteuid() != 0 ) 
    {
        printf( "\nerror: this program must be run as root\n\n" );
        return 1;
    }

    program = xdp_program__open_file( "server_xdp.o", "server_xdp", NULL );
    if ( libxdp_get_error( program ) ) 
    {
        printf( "\nerror: could not load server_xdp program\n\n");
        return 1;
    }

    printf( "after xdp_program__open_file\n" );
    fflush( stdout );

    int ret = xdp_program__attach( program, interface_index, XDP_MODE_NATIVE, 0 );
    if ( ret != 0 )
    {
        printf( "\nerror: failed to attach server_xdp program to interface\n\n" );
        return 1;
    }

    printf( "after xdp_program__attach\n" );
    fflush( stdout );

    int input_buffer_fd = bpf_obj_get( "/sys/fs/bpf/input_buffer" );
    if ( input_buffer_fd <= 0 )
    {
        printf( "\nerror: could not get input buffer: %s\n\n", strerror(errno) );
        return 1;
    }

    struct ring_buffer * input_buffer = ring_buffer__new( 0, process_input, NULL, NULL );
    if ( !input_buffer )
    {
        printf( "\nerror: could not create input ring buffer\n\n" );
        return 1;
    }

    printf( "cleaning up\n" );
    fflush( stdout );

    if ( program != NULL )
    {
        xdp_program__detach( program, interface_index, XDP_MODE_NATIVE, 0 );
        xdp_program__close( program );
    }

    return 0;
}
