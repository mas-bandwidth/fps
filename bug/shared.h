/*
    Shared definitions between XDP and userspace.
*/

#define JOIN_REQUEST_PACKET                                                                 1
#define JOIN_RESPONSE_PACKET                                                                2
#define INPUT_PACKET                                                                        3
#define STATS_REQUEST_PACKET                                                                4
#define STATS_RESPONSE_PACKET                                                               5

#define INPUT_SIZE                                                                        100
#define INPUTS_PER_PACKET                                                                  10
#define INPUT_PACKET_SIZE            ( 1 + 8 + 8 + 8 + (INPUT_SIZE + 8) * INPUTS_PER_PACKET )

#define PLAYER_DATA_SIZE                                                                 1024

#define JOIN_REQUEST_PACKET_SIZE                             ( 1 + 8 + 8 + PLAYER_DATA_SIZE )
#define JOIN_RESPONSE_PACKET_SIZE                                           ( 1 + 8 + 8 + 8 )
#define STATS_REQUEST_PACKET_SIZE                                                   ( 1 + 8 )
#define STATS_RESPONSE_PACKET_SIZE                                                  ( 1 + 8 )

#define MAX_SESSIONS                                                                  1000000

#define HEAP_SIZE                                                                        2048

#define PLAYER_STATE_SIZE                                                                1200

#define XDP_MAX_CPUS                                                                       16

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

struct stats_request_packet
{
    __u8 packet_type;
    __u64 inputs_processed;
};

struct server_stats
{
    __u64 inputs_processed;
};

struct session_data 
{
    __u64 next_input_sequence;
};

struct player_state
{
    __u64 t;
    __u8 data[PLAYER_STATE_SIZE];
};

struct input_header
{
    __u64 session_id;
    __u64 sequence;
    __u64 t;
};

struct input_data
{
    __u64 dt;
    __u8 input[INPUT_SIZE];
};

#pragma pack(pop)
