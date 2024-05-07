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
#include <sys/resource.h>
#include <sys/types.h>
#include <inttypes.h>
#include <time.h>
#include <pthread.h>
#include <sched.h>
#include <stdlib.h>

#define MAX_CPUS 32

volatile bool quit;

void clean_shutdown_handler( int signal )
{
    (void) signal; quit = true;
}

int main( int argc, char *argv[] )
{
    signal( SIGINT,  clean_shutdown_handler );
    signal( SIGTERM, clean_shutdown_handler );
    signal( SIGHUP,  clean_shutdown_handler );

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
            char * args[] = { "server", cpu_string, 0 };
            execv( "./worker", args );
            exit(0); 
        } 
    }

    // main loop

    while ( !quit )
    {
        usleep( 1000000 );
    }

    return 0;
}
