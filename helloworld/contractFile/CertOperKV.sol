// SPDX-License-Identifier: Apache-2.0
pragma solidity >=0.6.10 <0.8.20;

/// @title 极简版证书操作存储（KV 版本）
/// @notice 用 oper_id 作为 key，value 用字符串（推荐 JSON）保存整条记录
contract CertOperKV {
    // oper_id => JSON 文本（或你喜欢的任意字符串格式）
    mapping(bytes32 => string) private store;

    event Set(bytes32 indexed oper_id, string value);

    /// @notice 写入/覆盖一条操作记录
    /// @param oper_id 操作唯一标识（建议是 keccak256 或链下生成的 bytes32）
    /// @param value   建议用 JSON：例如
    ///  {"oper_type":0,"ca_pk":"...","cert_addr":"0x..","cert_target_state":0,
    ///   "latest_valid_oper":"0x..","pre_oper":"0x..","signature":"0x.."}
    function set(bytes32 oper_id, string memory value) public {
        store[oper_id] = value;
        emit Set(oper_id, value);
    }

    /// @notice 读取一条操作记录（原样返回）
    function get(bytes32 oper_id) public view returns (string memory) {
        return store[oper_id];
    }
}
