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

#include "shared.h"

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

// #define DEBUG 1

#if DEBUG
#define debug_printf bpf_printk
#else // #if DEBUG
#define debug_printf(...) do { } while (0)
#endif // #if DEBUG

struct {
    __uint( type, BPF_MAP_TYPE_LRU_PERCPU_HASH );
    __uint( map_flags, BPF_F_NO_COMMON_LRU );
    __type( key, __u64 );
    __type( value, struct session_data );
    __uint( max_entries, MAX_SESSIONS / MAX_CPUS );
    __uint( pinning, LIBBPF_PIN_BY_NAME );
} session_map SEC(".maps");

struct {
    __uint( type, BPF_MAP_TYPE_ARRAY );
    __uint( max_entries, 1 );
    __type( key, int );
    __type( value, struct server_stats );
    __uint( pinning, LIBBPF_PIN_BY_NAME );
} server_stats SEC(".maps");

struct {
    __uint( type, BPF_MAP_TYPE_PERF_EVENT_ARRAY );
    __uint( key_size, sizeof(int) );
    __uint( value_size, sizeof(int) );
    __uint( pinning, LIBBPF_PIN_BY_NAME );
} input_buffer SEC(".maps");

struct inner_player_state_map {
    __uint( type, BPF_MAP_TYPE_LRU_HASH );
    __type( key, __u64 );
    __type( value, struct player_state );
    __uint( max_entries, PLAYERS_PER_CPU );
} 
player_state_0 SEC(".maps"),
player_state_1 SEC(".maps"),
player_state_2 SEC(".maps"),
player_state_3 SEC(".maps"),
player_state_4 SEC(".maps"),
player_state_5 SEC(".maps"),
player_state_6 SEC(".maps"),
player_state_7 SEC(".maps"),
player_state_8 SEC(".maps"),
player_state_9 SEC(".maps"),
player_state_10 SEC(".maps"),
player_state_11 SEC(".maps"),
player_state_12 SEC(".maps"),
player_state_13 SEC(".maps"),
player_state_14 SEC(".maps"),
player_state_15 SEC(".maps"),
player_state_16 SEC(".maps"),
player_state_17 SEC(".maps"),
player_state_18 SEC(".maps"),
player_state_19 SEC(".maps"),
player_state_20 SEC(".maps"),
player_state_21 SEC(".maps"),
player_state_22 SEC(".maps"),
player_state_23 SEC(".maps"),
player_state_24 SEC(".maps"),
player_state_25 SEC(".maps"),
player_state_26 SEC(".maps"),
player_state_27 SEC(".maps"),
player_state_28 SEC(".maps"),
player_state_29 SEC(".maps"),
player_state_30 SEC(".maps"),
player_state_31 SEC(".maps");

struct {
    __uint( type, BPF_MAP_TYPE_ARRAY_OF_MAPS );
    __uint( max_entries, MAX_CPUS );
    __type( key, __u32 );
    __uint( pinning, LIBBPF_PIN_BY_NAME );
    __array( values, struct inner_player_state_map );
} player_state_map SEC(".maps") = {
    .values = { 
        &player_state_0,
        &player_state_1,
        &player_state_2,
        &player_state_3,
        &player_state_4,
        &player_state_5,
        &player_state_6,
        &player_state_7,
        &player_state_8,
        &player_state_9,
        &player_state_10,
        &player_state_11,
        &player_state_12,
        &player_state_13,
        &player_state_14,
        &player_state_15,
        &player_state_16,
        &player_state_17,
        &player_state_18,
        &player_state_19,
        &player_state_20,
        &player_state_21,
        &player_state_22,
        &player_state_23,
        &player_state_24,
        &player_state_25,
        &player_state_26,
        &player_state_27,
        &player_state_28,
        &player_state_29,
        &player_state_30,
        &player_state_31,
    }
};

struct heap {
    __u8 data[HEAP_SIZE];
};

struct {
    __uint( type, BPF_MAP_TYPE_PERCPU_ARRAY );
    __uint( max_entries, 1 );
    __type( key, int );
    __type( value, struct heap );
} heap_map SEC(".maps");

struct {
    __uint( type, BPF_MAP_TYPE_PERCPU_ARRAY );
    __uint( max_entries, 1 );
    __type( key, int );
    __type( value, struct counters );
    __uint( pinning, LIBBPF_PIN_BY_NAME );
} counters_map SEC(".maps");

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

                                debug_printf( "received packet type %d", packet_type );

                                if ( packet_type == JOIN_REQUEST_PACKET && (void*) payload + sizeof(struct join_request_packet) <= data_end )
                                {
                                    debug_printf( "received join request packet" );

                                    struct join_request_packet * request = (struct join_request_packet*) payload;

                                    struct session_data session;
                                    session.next_input_sequence = 1000;
                                    if ( bpf_map_update_elem( &session_map, &request->session_id, &session, BPF_NOEXIST ) == 0 )
                                    {
                                        debug_printf( "created session 0x%llx", request->session_id );
                                    }

                                    reflect_packet( data, sizeof(struct join_response_packet) );

                                    struct join_response_packet * response = (struct join_response_packet*) payload;

                                    response->packet_type = JOIN_RESPONSE_PACKET;
                                    response->server_time = get_server_time();

                                    bpf_xdp_adjust_tail( ctx, -( JOIN_REQUEST_PACKET_SIZE - JOIN_RESPONSE_PACKET_SIZE ) );

                                    return XDP_TX;
                                }
                                else if ( packet_type == INPUT_PACKET && (void*) payload + INPUT_PACKET_SIZE <= data_end )
                                {
                                    __u64 session_id = (__u64) payload[1];
                                    session_id |= ( (__u64) payload[2] ) << 8;
                                    session_id |= ( (__u64) payload[3] ) << 16;
                                    session_id |= ( (__u64) payload[4] ) << 24;
                                    session_id |= ( (__u64) payload[5] ) << 32;
                                    session_id |= ( (__u64) payload[6] ) << 40;
                                    session_id |= ( (__u64) payload[7] ) << 48;
                                    session_id |= ( (__u64) payload[8] ) << 56;

                                    struct session_data * session = (struct session_data*) bpf_map_lookup_elem( &session_map, &session_id );
                                    if ( session == NULL )
                                    {
                                        debug_printf( "could not find session 0x%llx", session_id );
                                        return XDP_DROP;
                                    }

                                    // send the input(s) down to userspace via perf buffer

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

                                    if ( sequence >= session->next_input_sequence )
                                    {
                                        __u64 n = ( sequence - session->next_input_sequence ) + 1;
                                        if ( n > 10 )
                                        {
                                            n = 10;
                                        }

                                        debug_printf( "process input %lld (n=%d)", sequence, n );

                                        session->next_input_sequence = sequence + 1;

                                        int zero = 0;
                                        __u8 * heap = (__u8*) bpf_map_lookup_elem( &heap_map, &zero );
                                        if ( !heap ) 
                                        {
                                            return XDP_DROP; // can't happen
                                        }

                                        if ( n == 1 && (void*) payload + 1 + 8 + 8 + 8 + ( 8 + INPUT_SIZE ) <= data_end )
                                        {
                                            for ( int i = 0; i < 8 + 8 + 8 + ( 8 + INPUT_SIZE ); i++ )
                                            {
                                                heap[i] = payload[1+i];
                                            }

                                            bpf_perf_event_output( ctx, &input_buffer, BPF_F_CURRENT_CPU, heap, 8 + 8 + 8 + ( 8 + INPUT_SIZE ) );
                                        }
                                        else if ( n == 2 && (void*) payload + 1 + 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 2 <= data_end )
                                        {
                                            for ( int i = 0; i < 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 2; i++ )
                                            {
                                                heap[i] = payload[1+i];
                                            }

                                            bpf_perf_event_output( ctx, &input_buffer, BPF_F_CURRENT_CPU, heap, 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 2 );
                                        }
                                        else if ( n == 3 && (void*) payload + 1 + 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 3 <= data_end )
                                        {
                                            for ( int i = 0; i < 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 3; i++ )
                                            {
                                                heap[i] = payload[1+i];
                                            }

                                            bpf_perf_event_output( ctx, &input_buffer, BPF_F_CURRENT_CPU, heap, 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 3 );
                                        }
                                        else if ( n == 4 && (void*) payload + 1 + 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 4 <= data_end )
                                        {
                                            for ( int i = 0; i < 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 4; i++ )
                                            {
                                                heap[i] = payload[1+i];
                                            }

                                            bpf_perf_event_output( ctx, &input_buffer, BPF_F_CURRENT_CPU, heap, 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 4 );
                                        }
                                        else if ( n == 5 && (void*) payload + 1 + 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 5 <= data_end )
                                        {
                                            for ( int i = 0; i < 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 5; i++ )
                                            {
                                                heap[i] = payload[1+i];
                                            }

                                            bpf_perf_event_output( ctx, &input_buffer, BPF_F_CURRENT_CPU, heap, 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 5 );
                                        }
                                        else if ( n == 6 && (void*) payload + 1 + 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 6 <= data_end )
                                        {
                                            for ( int i = 0; i < 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 6; i++ )
                                            {
                                                heap[i] = payload[1+i];
                                            }

                                            bpf_perf_event_output( ctx, &input_buffer, BPF_F_CURRENT_CPU, heap, 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 6 );
                                        }
                                        else if ( n == 7 && (void*) payload + 1 + 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 7 <= data_end )
                                        {
                                            for ( int i = 0; i < 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 7; i++ )
                                            {
                                                heap[i] = payload[1+i];
                                            }

                                            bpf_perf_event_output( ctx, &input_buffer, BPF_F_CURRENT_CPU, heap, 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 7 );
                                        }
                                        else if ( n == 8 && (void*) payload + 1 + 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 8 <= data_end )
                                        {
                                            for ( int i = 0; i < 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 8; i++ )
                                            {
                                                heap[i] = payload[1+i];
                                            }

                                            bpf_perf_event_output( ctx, &input_buffer, BPF_F_CURRENT_CPU, heap, 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 8 );
                                        }
                                        else if ( n == 9 && (void*) payload + 1 + 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 9 <= data_end )
                                        {
                                            for ( int i = 0; i < 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 9; i++ )
                                            {
                                                heap[i] = payload[1+i];
                                            }

                                            bpf_perf_event_output( ctx, &input_buffer, BPF_F_CURRENT_CPU, heap, 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 9 );
                                        }
                                        else if ( n == 10 && (void*) payload + 1 + 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 10 <= data_end )
                                        {
                                            for ( int i = 0; i < 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 10; i++ )
                                            {
                                                heap[i] = payload[1+i];
                                            }

                                            bpf_perf_event_output( ctx, &input_buffer, BPF_F_CURRENT_CPU, heap, 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 10 );
                                        }
                                    }
                                    else
                                    {
                                        debug_printf( "input packet is old" );
                                        return XDP_DROP;
                                    }

                                    // respond with a player state packet for the client's local player

                                    __u32 cpu = bpf_get_smp_processor_id();

                                    void * cpu_player_state_map = bpf_map_lookup_elem( &player_state_map, &cpu );
                                    if ( !cpu_player_state_map )
                                    {
                                        debug_printf( "could not find player state map for cpu %d", cpu );
                                        return XDP_DROP;
                                    }

                                    __u8 * player_state = (__u8*) bpf_map_lookup_elem( cpu_player_state_map, &session_id );
                                    if ( !player_state )
                                    {
                                        debug_printf( "couldn't find player state for session %x", session_id );
                                        return XDP_DROP;
                                    }

                                    // todo: this is breaking stuff?
                                    /*
                                    payload[0] = PLAYER_STATE_PACKET;

                                    for ( int i = 0; i < 8 + PLAYER_STATE_SIZE; i++ )
                                    {
                                        payload[1+i] = player_state[i];
                                    }
  
                                    int zero = 0;
                                    struct counters * counters = (struct counters*) bpf_map_lookup_elem( &counters_map, &zero );
                                    if ( !counters ) 
                                    {
                                        return XDP_DROP; // can't happen
                                    }

                                    __sync_fetch_and_add( &counters->player_state_packets_sent, 1 );

                                    debug_printf( "%d player state packets sent from cpu %d", counters->player_state_packets_sent, cpu );

                                    reflect_packet( data, PLAYER_STATE_PACKET_SIZE );

                                    bpf_xdp_adjust_tail( ctx, -( INPUT_PACKET_SIZE - PLAYER_STATE_PACKET_SIZE ) );

                                    return XDP_TX;
                                    */

                                }
                                else if ( packet_type == STATS_REQUEST_PACKET && (void*) payload + STATS_REQUEST_PACKET_SIZE <= data_end )
                                {
                                    debug_printf( "received stats request packet" );

                                    struct stats_request_packet * packet = (struct stats_request_packet*) payload;

                                    int zero = 0;
                                    struct server_stats * stats = (struct server_stats*) bpf_map_lookup_elem( &server_stats, &zero );
                                    if ( !stats ) 
                                    {
                                        return XDP_DROP; // can't happen
                                    }

                                    packet->packet_type = STATS_RESPONSE_PACKET;
                                    packet->inputs_processed = stats->inputs_processed;
                                    packet->player_state_packets_sent = stats->player_state_packets_sent;

                                    reflect_packet( data, sizeof(struct stats_request_packet) );

                                    return XDP_TX;
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
