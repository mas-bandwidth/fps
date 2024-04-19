/*
    FPS server XDP program

    USAGE:

        clang -Ilibbpf/src -g -O2 -target bpf -c server_xdp.c -o server_xdp.o
        sudo cat /sys/kernel/debug/tracing/trace_pipe
*/

#include <linux/in.h>
#include <linux/if_ether.h>
#include <linux/if_packet.h>
#include <linux/if_vlan.h>
#include <linux/ip.h>
#include <linux/ipv6.h>
#include <linux/udp.h>
#include <linux/bpf.h>
#include <linux/string.h>
#include <bpf/bpf_helpers.h>

#define SYNC_REQUEST_PACKET                                                                 1
#define SYNC_RESPONSE_PACKET                                                                2
#define INPUT_PACKET                                                                        3

#define INPUT_SIZE                                                                        100
#define INPUTS_PER_PACKET                                                                  10
#define INPUT_PACKET_SIZE                ( 1 + 8 + 8 + (INPUT_SIZE + 8) * INPUTS_PER_PACKET )

#if defined(__BYTE_ORDER__) && defined(__ORDER_LITTLE_ENDIAN__) && \
    __BYTE_ORDER__ == __ORDER_LITTLE_ENDIAN__
#define bpf_ntohs(x)        __builtin_bswap16(x)
#define bpf_htons(x)        __builtin_bswap16(x)
#elif defined(__BYTE_ORDER__) && defined(__ORDER_BIG_ENDIAN__) && \
    __BYTE_ORDER__ == __ORDER_BIG_ENDIAN__
#define bpf_ntohs(x)        (x)
#define bpf_htons(x)        (x)
#else
# error "Endianness detection needs to be set up for your compiler?!"
#endif

#define DEBUG 1

#if DEBUG
#define debug_printf bpf_printk
#else // #if DEBUG
#define debug_printf(...) do { } while (0)
#endif // #if DEBUG

SEC("server_xdp") int server_xdp_filter( struct xdp_md *ctx ) 
{ 
    void * data = (void*) (long) ctx->data; 

    void * data_end = (void*) (long) ctx->data_end; 

    struct ethhdr * eth = data;

    if ( (void*)eth + sizeof(struct ethhdr) < data_end )
    {
        if ( eth->h_proto == __constant_htons(ETH_P_IP) ) // IPV4
        {
            struct iphdr * ip = data + sizeof(struct ethhdr);

            if ( (void*)ip + sizeof(struct iphdr) < data_end )
            {
                if ( ip->protocol == IPPROTO_UDP ) // UDP
                {
                    struct udphdr * udp = (void*) ip + sizeof(struct iphdr);

                    if ( (void*)udp + sizeof(struct udphdr) <= data_end )
                    {
                        if ( udp->dest == __constant_htons(40000) )
                        {
                            __u8 * payload = (void*) udp + sizeof(struct udphdr);
                            int payload_bytes = data_end - (void*)payload;
                            if ( (void*)payload + 1 <= data_end )
                            {
                                int packet_type = payload[0];

                                debug_printf( "packet type is %d", packet_type );

                                if ( packet_type == INPUT_PACKET && (void*) payload + 1 + 8 + 8 + 8 + 8 + InputSize <= data_end )
                                {
                                    debug_printf( "received input packet" );

                                    __u64 session_id = (__u64) payload[1];
                                    session_id |= ( (__u64) payload[2] ) << 8;
                                    session_id |= ( (__u64) payload[3] ) << 16;
                                    session_id |= ( (__u64) payload[4] ) << 24;
                                    session_id |= ( (__u64) payload[5] ) << 32;
                                    session_id |= ( (__u64) payload[6] ) << 40;
                                    session_id |= ( (__u64) payload[7] ) << 48;
                                    session_id |= ( (__u64) payload[8] ) << 56;

                                    debug_printf( "session id = %016x", session_id );

                                    __u64 sequence = (__u64) payload[9];
                                    sequence |= ( (__u64) payload[10] ) << 8;
                                    sequence |= ( (__u64) payload[11] ) << 16;
                                    sequence |= ( (__u64) payload[12] ) << 24;
                                    sequence |= ( (__u64) payload[13] ) << 32;
                                    sequence |= ( (__u64) payload[14] ) << 40;
                                    sequence |= ( (__u64) payload[15] ) << 48;
                                    sequence |= ( (__u64) payload[16] ) << 56;

                                    debug_printf( "sequence = %lld", sequence );

                                    union double_uint64 {
                                        __u64 int_value;
                                        double float_value;
                                    };

                                    union double_uint64 t;
                                    t.int_value = (__u64) payload[17];
                                    t.int_value |= ( (__u64) payload[18] ) << 8;
                                    t.int_value |= ( (__u64) payload[19] ) << 16;
                                    t.int_value |= ( (__u64) payload[20] ) << 24;
                                    t.int_value |= ( (__u64) payload[21] ) << 32;
                                    t.int_value |= ( (__u64) payload[22] ) << 40;
                                    t.int_value |= ( (__u64) payload[23] ) << 48;
                                    t.int_value |= ( (__u64) payload[24] ) << 56;

                                    debug_printf( "t = %f", t );

                                    union double_uint64 dt;
                                    dt.int_value = (__u64) payload[25];
                                    dt.int_value |= ( (__u64) payload[26] ) << 8;
                                    dt.int_value |= ( (__u64) payload[27] ) << 16;
                                    dt.int_value |= ( (__u64) payload[28] ) << 24;
                                    dt.int_value |= ( (__u64) payload[29] ) << 32;
                                    dt.int_value |= ( (__u64) payload[30] ) << 40;
                                    dt.int_value |= ( (__u64) payload[31] ) << 48;
                                    dt.int_value |= ( (__u64) payload[32] ) << 56;

                                    debug_printf( "dt = %f", t );

                                    // todo: extract first input

                                    // todo: check if common case, eg. no packet loss, first input only

                                    // todo: else, handle going back n inputs
                                }
                                else
                                {
                                    debug_printf( "packet is too small (%d bytes)", payload_bytes );
                                }
                            }

                            return XDP_DROP;
                        }
                    }
                }
            }
        }
    }

    return XDP_PASS;
}

char _license[] SEC("license") = "GPL";
