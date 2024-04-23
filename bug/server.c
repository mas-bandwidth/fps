
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
    int interface_index;
    struct xdp_program * program;
    bool attached_native;
    bool attached_skb;
    int input_buffer_fd;

    if ( geteuid() != 0 ) 
    {
        printf( "\nerror: this program must be run as root\n\n" );
        return 1;
    }

    interface_index = 1;

    program = xdp_program__open_file( "server_xdp.o", "server_xdp", NULL );
    if ( libxdp_get_error( program ) ) 
    {
        printf( "\nerror: could not load server_xdp program\n\n");
        return 1;
    }

    int ret = xdp_program__attach( program, interface_index, XDP_MODE_NATIVE, 0 );
    if ( ret == 0 )
    {
        attached_native = true;
    } 
    else
    {
        printf( "falling back to skb mode...\n" );
        ret = xdp_program__attach( program, interface_index, XDP_MODE_SKB, 0 );
        if ( ret == 0 )
        {
            attached_skb = true;
        }
        else
        {
            printf( "\nerror: failed to attach server_xdp program to interface\n\n" );
            return 1;
        }
    }

    input_buffer_fd = 0;

    struct ring_buffer * input_buffer = ring_buffer__new( input_buffer_fd, process_input, NULL, NULL );

    if ( program != NULL )
    {
        if ( attached_native )
        {
            xdp_program__detach( program, interface_index, XDP_MODE_NATIVE, 0 );
        }
        if ( attached_skb )
        {
            xdp_program__detach( program, interface_index, XDP_MODE_SKB, 0 );
        }
        xdp_program__close( program );
    }

    return 0;
}
