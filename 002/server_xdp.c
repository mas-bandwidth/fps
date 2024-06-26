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

#define DEBUG 1

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
    __uint( type, BPF_MAP_TYPE_PERF_EVENT_ARRAY );
    __uint( key_size, sizeof(int) );
    __uint( value_size, sizeof(int) );
    __uint( pinning, LIBBPF_PIN_BY_NAME );
} input_buffer SEC(".maps");

struct heap {
    __u8 data[HEAP_SIZE];
};

struct {
    __uint( type, BPF_MAP_TYPE_PERCPU_ARRAY );
    __uint( max_entries, 1 );
    __type( key, int );
    __type( value, struct heap );
} heap SEC(".maps");

struct {
    __uint( type, BPF_MAP_TYPE_ARRAY );
    __uint( max_entries, 1 );
    __type( key, int );
    __type( value, struct server_stats );
    __uint( pinning, LIBBPF_PIN_BY_NAME );
} server_stats SEC(".maps");

struct inner_player_state_map {
    __uint( type, BPF_MAP_TYPE_LRU_HASH );
    __type( key, __u64 );
    __type( value, struct player_state );
    __uint( max_entries, MAX_SESSIONS / MAX_CPUS );
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
player_state_31 SEC(".maps"),
player_state_32 SEC(".maps"),
player_state_33 SEC(".maps"),
player_state_34 SEC(".maps"),
player_state_35 SEC(".maps"),
player_state_36 SEC(".maps"),
player_state_37 SEC(".maps"),
player_state_38 SEC(".maps"),
player_state_39 SEC(".maps"),
player_state_40 SEC(".maps"),
player_state_41 SEC(".maps"),
player_state_42 SEC(".maps"),
player_state_43 SEC(".maps"),
player_state_44 SEC(".maps"),
player_state_45 SEC(".maps"),
player_state_46 SEC(".maps"),
player_state_47 SEC(".maps"),
player_state_48 SEC(".maps"),
player_state_49 SEC(".maps"),
player_state_50 SEC(".maps"),
player_state_51 SEC(".maps"),
player_state_52 SEC(".maps"),
player_state_53 SEC(".maps"),
player_state_54 SEC(".maps"),
player_state_55 SEC(".maps"),
player_state_56 SEC(".maps"),
player_state_57 SEC(".maps"),
player_state_58 SEC(".maps"),
player_state_59 SEC(".maps"),
player_state_60 SEC(".maps"),
player_state_61 SEC(".maps"),
player_state_62 SEC(".maps"),
player_state_63 SEC(".maps"),
player_state_64 SEC(".maps"),
player_state_65 SEC(".maps"),
player_state_66 SEC(".maps"),
player_state_67 SEC(".maps"),
player_state_68 SEC(".maps"),
player_state_69 SEC(".maps"),
player_state_70 SEC(".maps"),
player_state_71 SEC(".maps"),
player_state_72 SEC(".maps"),
player_state_73 SEC(".maps"),
player_state_74 SEC(".maps"),
player_state_75 SEC(".maps"),
player_state_76 SEC(".maps"),
player_state_77 SEC(".maps"),
player_state_78 SEC(".maps"),
player_state_79 SEC(".maps"),
player_state_80 SEC(".maps"),
player_state_81 SEC(".maps"),
player_state_82 SEC(".maps"),
player_state_83 SEC(".maps"),
player_state_84 SEC(".maps"),
player_state_85 SEC(".maps"),
player_state_86 SEC(".maps"),
player_state_87 SEC(".maps"),
player_state_88 SEC(".maps"),
player_state_89 SEC(".maps"),
player_state_90 SEC(".maps"),
player_state_91 SEC(".maps"),
player_state_92 SEC(".maps"),
player_state_93 SEC(".maps"),
player_state_94 SEC(".maps"),
player_state_95 SEC(".maps"),
player_state_96 SEC(".maps"),
player_state_97 SEC(".maps"),
player_state_98 SEC(".maps"),
player_state_99 SEC(".maps"),
player_state_100 SEC(".maps"),
player_state_101 SEC(".maps"),
player_state_102 SEC(".maps"),
player_state_103 SEC(".maps"),
player_state_104 SEC(".maps"),
player_state_105 SEC(".maps"),
player_state_106 SEC(".maps"),
player_state_107 SEC(".maps"),
player_state_108 SEC(".maps"),
player_state_109 SEC(".maps"),
player_state_110 SEC(".maps"),
player_state_111 SEC(".maps"),
player_state_112 SEC(".maps"),
player_state_113 SEC(".maps"),
player_state_114 SEC(".maps"),
player_state_115 SEC(".maps"),
player_state_116 SEC(".maps"),
player_state_117 SEC(".maps"),
player_state_118 SEC(".maps"),
player_state_119 SEC(".maps"),
player_state_120 SEC(".maps"),
player_state_121 SEC(".maps"),
player_state_122 SEC(".maps"),
player_state_123 SEC(".maps"),
player_state_124 SEC(".maps"),
player_state_125 SEC(".maps"),
player_state_126 SEC(".maps"),
player_state_127 SEC(".maps");

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
        &player_state_32,
        &player_state_33,
        &player_state_34,
        &player_state_35,
        &player_state_36,
        &player_state_37,
        &player_state_38,
        &player_state_39,
        &player_state_40,
        &player_state_41,
        &player_state_42,
        &player_state_43,
        &player_state_44,
        &player_state_45,
        &player_state_46,
        &player_state_47,
        &player_state_48,
        &player_state_49,
        &player_state_50,
        &player_state_51,
        &player_state_52,
        &player_state_53,
        &player_state_54,
        &player_state_55,
        &player_state_56,
        &player_state_57,
        &player_state_58,
        &player_state_59,
        &player_state_60,
        &player_state_61,
        &player_state_62,
        &player_state_63,
        &player_state_64,
        &player_state_65,
        &player_state_66,
        &player_state_67,
        &player_state_68,
        &player_state_69,
        &player_state_70,
        &player_state_71,
        &player_state_72,
        &player_state_73,
        &player_state_74,
        &player_state_75,
        &player_state_76,
        &player_state_77,
        &player_state_78,
        &player_state_79,
        &player_state_80,
        &player_state_81,
        &player_state_82,
        &player_state_83,
        &player_state_84,
        &player_state_85,
        &player_state_86,
        &player_state_87,
        &player_state_88,
        &player_state_89,
        &player_state_90,
        &player_state_91,
        &player_state_92,
        &player_state_93,
        &player_state_94,
        &player_state_95,
        &player_state_96,
        &player_state_97,
        &player_state_98,
        &player_state_99,
        &player_state_100,
        &player_state_101,
        &player_state_102,
        &player_state_103,
        &player_state_104,
        &player_state_105,
        &player_state_106,
        &player_state_107,
        &player_state_108,
        &player_state_109,
        &player_state_110,
        &player_state_111,
        &player_state_112,
        &player_state_113,
        &player_state_114,
        &player_state_115,
        &player_state_116,
        &player_state_117,
        &player_state_118,
        &player_state_119,
        &player_state_120,
        &player_state_121,
        &player_state_122,
        &player_state_123,
        &player_state_124,
        &player_state_125,
        &player_state_126,
        &player_state_127,
    }
};

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
                                else if ( packet_type == INPUT_PACKET && (void*) payload + 1 + 8 + 8 + 8 + 8 + INPUT_SIZE <= data_end )
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
                                        __u8 * data = (__u8*) bpf_map_lookup_elem( &heap, &zero );
                                        if ( !data ) 
                                        {
                                            return XDP_DROP; // can't happen
                                        }

                                        if ( n == 1 && (void*) payload + 1 + 8 + 8 + 8 + ( 8 + INPUT_SIZE ) <= data_end )
                                        {
                                            for ( int i = 0; i < 8 + 8 + 8 + ( 8 + INPUT_SIZE ); i++ )
                                            {
                                                data[i] = payload[1+i];
                                            }

                                            bpf_perf_event_output( ctx, &input_buffer, BPF_F_CURRENT_CPU, data, 8 + 8 + 8 + ( 8 + INPUT_SIZE ) );
                                        }
                                        else if ( n == 2 && (void*) payload + 1 + 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 2 <= data_end )
                                        {
                                            for ( int i = 0; i < 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 2; i++ )
                                            {
                                                data[i] = payload[1+i];
                                            }

                                            bpf_perf_event_output( ctx, &input_buffer, BPF_F_CURRENT_CPU, data, 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 2 );
                                        }
                                        else if ( n == 3 && (void*) payload + 1 + 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 3 <= data_end )
                                        {
                                            for ( int i = 0; i < 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 3; i++ )
                                            {
                                                data[i] = payload[1+i];
                                            }

                                            bpf_perf_event_output( ctx, &input_buffer, BPF_F_CURRENT_CPU, data, 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 3 );
                                        }
                                        else if ( n == 4 && (void*) payload + 1 + 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 4 <= data_end )
                                        {
                                            for ( int i = 0; i < 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 4; i++ )
                                            {
                                                data[i] = payload[1+i];
                                            }

                                            bpf_perf_event_output( ctx, &input_buffer, BPF_F_CURRENT_CPU, data, 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 4 );
                                        }
                                        else if ( n == 5 && (void*) payload + 1 + 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 5 <= data_end )
                                        {
                                            for ( int i = 0; i < 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 5; i++ )
                                            {
                                                data[i] = payload[1+i];
                                            }

                                            bpf_perf_event_output( ctx, &input_buffer, BPF_F_CURRENT_CPU, data, 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 5 );
                                        }
                                        else if ( n == 6 && (void*) payload + 1 + 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 6 <= data_end )
                                        {
                                            for ( int i = 0; i < 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 6; i++ )
                                            {
                                                data[i] = payload[1+i];
                                            }

                                            bpf_perf_event_output( ctx, &input_buffer, BPF_F_CURRENT_CPU, data, 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 6 );
                                        }
                                        else if ( n == 7 && (void*) payload + 1 + 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 7 <= data_end )
                                        {
                                            for ( int i = 0; i < 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 7; i++ )
                                            {
                                                data[i] = payload[1+i];
                                            }

                                            bpf_perf_event_output( ctx, &input_buffer, BPF_F_CURRENT_CPU, data, 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 7 );
                                        }
                                        else if ( n == 8 && (void*) payload + 1 + 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 8 <= data_end )
                                        {
                                            for ( int i = 0; i < 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 8; i++ )
                                            {
                                                data[i] = payload[1+i];
                                            }

                                            bpf_perf_event_output( ctx, &input_buffer, BPF_F_CURRENT_CPU, data, 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 8 );
                                        }
                                        else if ( n == 9 && (void*) payload + 1 + 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 9 <= data_end )
                                        {
                                            for ( int i = 0; i < 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 9; i++ )
                                            {
                                                data[i] = payload[1+i];
                                            }

                                            bpf_perf_event_output( ctx, &input_buffer, BPF_F_CURRENT_CPU, data, 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 9 );
                                        }
                                        else if ( n == 10 && (void*) payload + 1 + 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 10 <= data_end )
                                        {
                                            for ( int i = 0; i < 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 10; i++ )
                                            {
                                                data[i] = payload[1+i];
                                            }

                                            bpf_perf_event_output( ctx, &input_buffer, BPF_F_CURRENT_CPU, data, 8 + 8 + 8 + ( 8 + INPUT_SIZE ) * 10 );
                                        }
                                    }
                                    else
                                    {
                                        debug_printf( "input packet is old" );
                                    }
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
