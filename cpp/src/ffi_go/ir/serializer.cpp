#include "serializer.h"

#include <iostream>
#include <cstdint>
#include <memory>
#include <string>
#include <string_view>
#include <vector>

#include <msgpack.hpp>

#include <clp/components/core/src/clp/ffi/encoding_methods.hpp>
#include <clp/components/core/src/clp/ffi/ir_stream/decoding_methods.hpp>
#include <clp/components/core/src/clp/ffi/ir_stream/encoding_methods.hpp>
#include <clp/components/core/src/clp/ir/types.hpp>
#include <clp/components/core/src/clp_s/ffi/ir_stream/serialization_methods.hpp>

#include <ffi_go/defs.h>
#include <ffi_go/ir/LogTypes.hpp>
#include <ffi_go/LogTypes.hpp>

namespace ffi_go::ir {
using namespace clp::ffi::ir_stream;
using clp::ir::eight_byte_encoded_variable_t;
using clp::ir::four_byte_encoded_variable_t;
using clp_s::ffi::ir_stream::serialize_key_value_pair_record;

namespace {
    /**
     * Generic helper for ir_serializer_serialize_*_log_event
     */
    template <class encoded_variable_t>
    auto serialize_log_event(
            StringView log_message,
            epoch_time_ms_t timestamp_or_delta,
            void* ir_serializer,
            ByteSpan* ir_view
    ) -> int {
        Serializer* serializer{static_cast<Serializer*>(ir_serializer)};
        serializer->m_ir_buf.clear();

        bool success{false};
        if constexpr (std::is_same_v<encoded_variable_t, eight_byte_encoded_variable_t>) {
            success = eight_byte_encoding::serialize_log_event(
                    timestamp_or_delta,
                    std::string_view{log_message.m_data, log_message.m_size},
                    serializer->m_logtype,
                    serializer->m_ir_buf
            );
        } else if constexpr (std::is_same_v<encoded_variable_t, four_byte_encoded_variable_t>) {
            success = four_byte_encoding::serialize_log_event(
                    timestamp_or_delta,
                    std::string_view{log_message.m_data, log_message.m_size},
                    serializer->m_logtype,
                    serializer->m_ir_buf
            );
        } else {
            static_assert(cAlwaysFalse<encoded_variable_t>, "Invalid/unhandled encoding type");
        }
        if (false == success) {
            std::cout << 2 << std::endl;
            return static_cast<int>(IRErrorCode_Corrupted_IR);
        }

        ir_view->m_data = serializer->m_ir_buf.data();
        ir_view->m_size = serializer->m_ir_buf.size();
        return static_cast<int>(IRErrorCode_Success);
    }
}  // namespace

extern "C" auto ir_serializer_close(void* ir_serializer) -> void {
    // NOLINTNEXTLINE(cppcoreguidelines-owning-memory)
    delete static_cast<Serializer*>(ir_serializer);
}

extern "C" auto ir_serializer_serialize_eight_byte_preamble(
        StringView ts_pattern,
        StringView ts_pattern_syntax,
        StringView time_zone_id,
        void** ir_serializer_ptr,
        ByteSpan* ir_view
) -> int {
    Serializer* serializer{new Serializer{}};
    *ir_serializer_ptr = serializer;
    if (false
        == eight_byte_encoding::serialize_preamble(
                std::string_view{ts_pattern.m_data, ts_pattern.m_size},
                std::string_view{ts_pattern_syntax.m_data, ts_pattern_syntax.m_size},
                std::string_view{time_zone_id.m_data, time_zone_id.m_size},
                serializer->m_ir_buf
        ))
    {
        std::cout << 3 << std::endl;
        return static_cast<int>(IRErrorCode_Corrupted_IR);
    }

    ir_view->m_data = serializer->m_ir_buf.data();
    ir_view->m_size = serializer->m_ir_buf.size();
    return static_cast<int>(IRErrorCode_Success);
}

extern "C" auto ir_serializer_serialize_four_byte_preamble(
        StringView ts_pattern,
        StringView ts_pattern_syntax,
        StringView time_zone_id,
        epoch_time_ms_t reference_ts,
        void** ir_serializer_ptr,
        ByteSpan* ir_view
) -> int {
    if (nullptr == ir_serializer_ptr || nullptr == ir_view) {
        std::cout << 4 << std::endl;
        return static_cast<int>(IRErrorCode_Corrupted_IR);
    }
    Serializer* serializer{new Serializer{}};
    if (nullptr == serializer) {
        std::cout << 5 << std::endl;
        return static_cast<int>(IRErrorCode_Corrupted_IR);
    }
    *ir_serializer_ptr = serializer;
    if (false
        == four_byte_encoding::serialize_preamble(
                std::string_view{ts_pattern.m_data, ts_pattern.m_size},
                std::string_view{ts_pattern_syntax.m_data, ts_pattern_syntax.m_size},
                std::string_view{time_zone_id.m_data, time_zone_id.m_size},
                reference_ts,
                serializer->m_ir_buf
        ))
    {
        std::cout << 6 << std::endl;
        return static_cast<int>(IRErrorCode_Corrupted_IR);
    }

    ir_view->m_data = serializer->m_ir_buf.data();
    ir_view->m_size = serializer->m_ir_buf.size();
    return static_cast<int>(IRErrorCode_Success);
}

extern "C" auto ir_serializer_serialize_eight_byte_log_event(
        StringView log_message,
        epoch_time_ms_t timestamp,
        void* ir_serializer,
        ByteSpan* ir_view
) -> int {
    return serialize_log_event<eight_byte_encoded_variable_t>(
            log_message,
            timestamp,
            ir_serializer,
            ir_view
    );
}

extern "C" auto ir_serializer_serialize_four_byte_log_event(
        StringView log_message,
        epoch_time_ms_t timestamp_delta,
        void* ir_serializer,
        ByteSpan* ir_view
) -> int {
    return serialize_log_event<four_byte_encoded_variable_t>(
            log_message,
            timestamp_delta,
            ir_serializer,
            ir_view
    );
}

extern "C" auto ir_kv_serializer_close(void* ir_kv_serializer) -> void {
    delete static_cast<KVSerializer*>(ir_kv_serializer);
}

extern "C" auto ir_serializer_serialize_kv_preamble(
        void** ir_kv_serializer_ptr,
        ByteSpan* ir_view
) -> int {
    if (nullptr == ir_kv_serializer_ptr || nullptr == ir_view) {
        std::cout << 7 << std::endl;
        return static_cast<int>(IRErrorCode_Corrupted_IR);
    }
    auto* kv_serializer{new KVSerializer{}};
    *ir_kv_serializer_ptr = kv_serializer;
    auto const buf_view{kv_serializer->m_serialization_buffer.get_ir_buf()};
    ir_view->m_data = static_cast<void*>(const_cast<char*>(buf_view.data()));
    ir_view->m_size = buf_view.size();
    return static_cast<int>(IRErrorCode_Success);
}

extern "C" auto ir_serializer_serialize_kv_log_event(
        ByteSpan msgpack_bytes,
        void* ir_kv_serializer,
        ByteSpan* ir_view
) -> int {
    if (nullptr == ir_kv_serializer || nullptr == ir_view) {
        std::cout << 8 << std::endl;
        return static_cast<int>(IRErrorCode_Corrupted_IR);
    }
    auto* serializer{static_cast<KVSerializer*>(ir_kv_serializer)};
    serializer->m_serialization_buffer.flush_ir_buf();
    msgpack::object_handle oh;
    msgpack::unpack(oh, static_cast<char const*>(msgpack_bytes.m_data), msgpack_bytes.m_size);
    if (false == serialize_key_value_pair_record(oh.get(), serializer->m_serialization_buffer)) {
        std::cout << 9 << std::endl;
        return static_cast<int>(IRErrorCode_Corrupted_IR);
    }
    auto const buf_view{serializer->m_serialization_buffer.get_ir_buf()};
    ir_view->m_data = static_cast<void*>(const_cast<char*>(buf_view.data()));
    ir_view->m_size = buf_view.size();
    return static_cast<int>(IRErrorCode_Success);
}
}  // namespace ffi_go::ir
