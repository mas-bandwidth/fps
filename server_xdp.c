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

#define JOIN_REQUEST_PACKET                                                                 1
#define JOIN_RESPONSE_PACKET                                                                2
#define INPUT_PACKET                                                                        3

#define INPUT_SIZE                                                                        100
#define INPUTS_PER_PACKET                                                                  10
#define INPUT_PACKET_SIZE                ( 1 + 8 + 8 + (INPUT_SIZE + 8) * INPUTS_PER_PACKET )

#define PLAYER_DATA_SIZE                                                                 1024

#define JOIN_REQUEST_PACKET_SIZE                             ( 1 + 8 + 8 + PLAYER_DATA_SIZE )
#define JOIN_RESPONSE_PACKET_SIZE                                           ( 1 + 8 + 8 + 8 )

#define MAX_SESSIONS                                                                  1000000

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

#pragma pack(push, 1)

struct join_request_packet
{
    __u8 packet_type;
    __u64 session_id;
    __u64 send_time;
    __u8 player_data[PLAYER_DATA_SIZE];
};

struct join_response_packet
{
    __u8 packet_type;
    __u64 session_id;
    __u64 send_time;
    __u64 server_time;
};

#pragma pack(pop)

struct session_data 
{
    __u8 player_data[PLAYER_DATA_SIZE];
    __u64 last_input_sequence;
};

struct {
    __uint( type, BPF_MAP_TYPE_LRU_HASH );
    __type( key, __u64 );
    __type( value, __u64 );
    __uint( max_entries, MAX_SESSIONS );
    __uint( pinning, LIBBPF_PIN_BY_NAME );
} session_map SEC(".maps");

static void reflect_packet( void * data, int payload_bytes )
{
    struct ethhdr * eth = data;
    struct iphdr  * ip  = data + sizeof( struct ethhdr );
    struct udphdr * udp = (void*) ip + sizeof( struct iphdr );

    __u16 a = udp->source;
    udp->source = udp->dest;
    udp->dest = a;
    udp->check = 0;
    udp->len = bpf_htons( sizeof(struct udphdr) + payload_bytes );

    __u32 b = ip->saddr;
    ip->saddr = ip->daddr;
    ip->daddr = b;
    ip->tot_len = bpf_htons( sizeof(struct iphdr) + sizeof(struct udphdr) + payload_bytes );
    ip->check = 0;

    char c[ETH_ALEN];
    memcpy( c, eth->h_source, ETH_ALEN );
    memcpy( eth->h_source, eth->h_dest, ETH_ALEN );
    memcpy( eth->h_dest, c, ETH_ALEN );

    __u16 * p = (__u16*) ip;
    __u32 checksum = p[0];
    checksum += p[1];
    checksum += p[2];
    checksum += p[3];
    checksum += p[4];
    checksum += p[5];
    checksum += p[6];
    checksum += p[7];
    checksum += p[8];
    checksum += p[9];
    checksum = ~ ( ( checksum & 0xFFFF ) + ( checksum >> 16 ) );
    ip->check = checksum;
}

static __u64 get_server_time()
{
    return bpf_ktime_get_boot_ns();
}

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

                                if ( packet_type == JOIN_REQUEST_PACKET && (void*) payload + sizeof(struct join_request_packet) <= data_end )
                                {
                                    debug_printf( "received join request packet" );

                                    struct join_request_packet * request = (struct join_request_packet*) payload;

                                    // todo: create map entry for session id if it doesn't already exist

                                    reflect_packet( data, sizeof(struct join_response_packet) );

                                    struct join_response_packet * response = (struct join_response_packet*) payload;

                                    response->packet_type = JOIN_RESPONSE_PACKET;
                                    response->server_time = get_server_time();

                                    bpf_xdp_adjust_tail( ctx, -( JOIN_REQUEST_PACKET_SIZE - JOIN_RESPONSE_PACKET_SIZE ) );

                                    return XDP_TX;
                                }
                                else if ( packet_type == INPUT_PACKET && (void*) payload + 1 + 8 + 8 + 8 + 8 + INPUT_SIZE <= data_end )
                                {
                                    debug_printf( "received input packet" );

                                    // todo: switch to this via overlay struct

                                    __u64 session_id = (__u64) payload[1];
                                    session_id |= ( (__u64) payload[2] ) << 8;
                                    session_id |= ( (__u64) payload[3] ) << 16;
                                    session_id |= ( (__u64) payload[4] ) << 24;
                                    session_id |= ( (__u64) payload[5] ) << 32;
                                    session_id |= ( (__u64) payload[6] ) << 40;
                                    session_id |= ( (__u64) payload[7] ) << 48;
                                    session_id |= ( (__u64) payload[8] ) << 56;

                                    __u64 sequence = (__u64) payload[9];
                                    sequence |= ( (__u64) payload[10] ) << 8;
                                    sequence |= ( (__u64) payload[11] ) << 16;
                                    sequence |= ( (__u64) payload[12] ) << 24;
                                    sequence |= ( (__u64) payload[13] ) << 32;
                                    sequence |= ( (__u64) payload[14] ) << 40;
                                    sequence |= ( (__u64) payload[15] ) << 48;
                                    sequence |= ( (__u64) payload[16] ) << 56;

                                    __u64 t = (__u64) payload[17];
                                    t |= ( (__u64) payload[18] ) << 8;
                                    t |= ( (__u64) payload[19] ) << 16;
                                    t |= ( (__u64) payload[20] ) << 24;
                                    t |= ( (__u64) payload[21] ) << 32;
                                    t |= ( (__u64) payload[22] ) << 40;
                                    t |= ( (__u64) payload[23] ) << 48;
                                    t |= ( (__u64) payload[24] ) << 56;

                                    __u64 dt = (__u64) payload[25];
                                    dt |= ( (__u64) payload[26] ) << 8;
                                    dt |= ( (__u64) payload[27] ) << 16;
                                    dt |= ( (__u64) payload[28] ) << 24;
                                    dt |= ( (__u64) payload[29] ) << 32;
                                    dt |= ( (__u64) payload[30] ) << 40;
                                    dt |= ( (__u64) payload[31] ) << 48;
                                    dt |= ( (__u64) payload[32] ) << 56;

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
