
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
    int interface_index = 1;

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

    printf( "xdp_program__open_file\n" );
    fflush( stdout );

    int ret = xdp_program__attach( program, interface_index, XDP_MODE_NATIVE, 0 );
    if ( ret != 0 )
    {
        printf( "\nerror: failed to attach server_xdp program to interface\n\n" );
        return 1;
    }

    printf( "xdp_program__attach\n" );
    fflush( stdout );

    struct ring_buffer * input_buffer = ring_buffer__new( 0, process_input, NULL, NULL );

    printf( "cleaning up\n" );
    fflush( stdout );

    if ( program != NULL )
    {
        xdp_program__detach( program, interface_index, XDP_MODE_NATIVE, 0 );
        xdp_program__close( program );
    }

    return 0;
}
