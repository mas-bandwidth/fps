
#include "shared.h"

#define MAP_BUCKET_SIZE                          32
#define MAP_NUM_BUCKETS             PLAYERS_PER_CPU

struct map_element_t
{
    uint64_t session_id;
    uint8_t * player_data;
};

struct map_bucket_t
{
    int size;
    struct map_element_t elements[MAP_BUCKET_SIZE];
};

struct map_t
{
    int size;
    struct map_bucket_t buckets[MAP_NUM_BUCKETS];
};

static void map_reset( struct map_t * map );

struct map_t * map_create()
{
    struct map_t * map = (struct map_t*) malloc( sizeof( struct map_t ) );
    assert( map );
    memset( map, 0, sizeof(map) );
    map_reset( map );
    return map;
}

static void map_destroy( struct map_t * map )
{
    assert( map );
    free( map );
}

static void map_bucket_reset( struct map_bucket_t * bucket )
{
    assert( bucket );
    bucket->size = 0;
    for ( int i = 0; i < MAP_NUM_BUCKETS; i++ )
    {
        struct map_element_t * element = bucket->elements + i;
        element->session_id = 0;
        if ( element->player_data )
        {
            free( element->player_data );
            element->player_data = NULL;
        }
    }
}

static void map_reset( struct map_t * map )
{
    assert( map );
    map->size = 0;
    for ( int i = 0; i < MAP_NUM_BUCKETS; i++ )
    {
        struct map_bucket_t * bucket = map->buckets + i;
        map_bucket_reset( bucket );
    }
}

static int map_set( struct map_t * map, uint64_t session_id, void * player_data )
{
    int bucket_index = session_id % MAP_NUM_BUCKETS;
    struct map_bucket_t * bucket = map->buckets + bucket_index;
    if ( bucket->size == MAP_BUCKET_SIZE )
    {
        return 0;
    }

    struct map_element_t * element = bucket->elements + bucket->size;
    element->session_id = session_id;
    element->player_data = player_data;

    ++bucket->size;
    ++map->size;

    return 1;
}

static struct map_element_t * map_bucket_find( struct map_bucket_t * bucket, uint64_t session_id )
{
    for ( int i = 0; i < bucket->size; i++ )
    {
        if ( bucket->elements[i].session_id == session_id )
        {
            return &bucket->elements[i];
        }
    }
    return NULL;
}

static void * map_get( struct map_t * map, uint64_t session_id )
{
    int bucket_index = session_id % MAP_NUM_BUCKETS;
    struct map_bucket_t * bucket = map->buckets + bucket_index;
    struct map_element_t * element = map_bucket_find( bucket, session_id );
    return element ? element->player_data : NULL;
}

static int map_delete( struct map_t * map, uint64_t session_id )
{
    int bucket_index = session_id % MAP_NUM_BUCKETS;
    struct netcode_address_map_bucket_t * bucket = map->buckets + bucket_index;
    struct netcode_address_map_element_t * element = map_bucket_find( bucket, session_id );
    if ( !element )
    {
        return 0;
    }

    struct map_element_t * last = bucket->elements + ( bucket->size - 1 );
    *element = *last;
    last->session_id = 0;
    if ( last->player_data )
    {
        free( last->player_data );
        last->player_data = NULL;    
    }
    
    --bucket->size;
    --map->size;

    return 1;
}
