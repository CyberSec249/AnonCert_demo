// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.6.10;

/// @title DPCE Certificate Transparency Registry for FISCO-BCOS
/// @notice 记录证书生命周期的每一次操作（签发/更新/撤销），提供双指针：
///         - pre_oper：上一条操作（不区分有效/类型）
///         - latest_valid_oper：上一条“有效状态”的操作
///         便于审计及撤销回滚分析。
contract CertOper {
    struct DPCEOper {
        bytes32  oper_id;            // 操作唯一标识
        uint8    oper_type;          // 0: issuance; 1: update; 2: revocation
        string   ca_pk;              // 操作CA的公钥（建议PEM/HEX/DER文本之一）
        bytes32  cert_addr;          // 证书地址/定位符（自行约定，可用 keccak(pubkey)）
        uint8    cert_target_state;  // 0: valid; 1: invalid
        bytes32  latest_valid_oper;  // 指向上一个“有效状态”的操作
        bytes32  pre_oper;           // 指向紧邻前一操作
        bytes    signature;          // CA对本次操作的签名（格式自定义，链上不强校验）
        uint256  blockTime;          // 上链时间（链上记录）
        address  sender;             // 交易发起者
    }

    // oper_id => 记录
    mapping(bytes32 => DPCEOper) private ops;
    // oper_id 是否已存在
    mapping(bytes32 => bool)     private opExists;
    // cert_addr => 最新操作 oper_id
    mapping(bytes32 => bytes32)  private latestByCert;

    event CertOperSet(
        bytes32 indexed oper_id,
        bytes32 indexed cert_addr,
        uint8   oper_type,
        uint8   cert_target_state,
        bytes32 latest_valid_oper,
        bytes32 pre_oper,
        address indexed sender
    );

    /// @notice 写入一条证书操作记录（上链）
    function setCertOper(
        bytes32 oper_id,
        uint8   oper_type,          // 0/1/2
        string  calldata ca_pk,
        bytes32 cert_addr,
        uint8   cert_target_state,  // 0/1
        bytes32 latest_valid_oper,
        bytes32 pre_oper,
        bytes   calldata signature
    ) external {
        require(oper_id != bytes32(0), "oper_id required");
        require(!opExists[oper_id], "oper_id exists");
        require(oper_type <= 2, "bad oper_type");
        require(cert_target_state <= 1, "bad target_state");
        require(cert_addr != bytes32(0), "cert_addr required");

        // 指针（如果提供）必须已存在
        if (latest_valid_oper != bytes32(0)) {
            require(opExists[latest_valid_oper], "latest_valid_oper not found");
        }
        if (pre_oper != bytes32(0)) {
            require(opExists[pre_oper], "pre_oper not found");
        }

        DPCEOper storage o = ops[oper_id];
        o.oper_id            = oper_id;
        o.oper_type          = oper_type;
        o.ca_pk              = ca_pk;
        o.cert_addr          = cert_addr;
        o.cert_target_state  = cert_target_state;
        o.latest_valid_oper  = latest_valid_oper;
        o.pre_oper           = pre_oper;
        o.signature          = signature;
        o.blockTime          = block.timestamp;
        o.sender             = msg.sender;

        opExists[oper_id] = true;
        latestByCert[cert_addr] = oper_id; // 新记录成为该证书的最新节点

        emit CertOperSet(
            oper_id, cert_addr, oper_type, cert_target_state,
            latest_valid_oper, pre_oper, msg.sender
        );
    }

    /// @notice 按 oper_id 查询一条记录
    function getCertOper(bytes32 oper_id) external view returns (
        bytes32  _oper_id,
        uint8    _oper_type,
        string  memory _ca_pk,
        bytes32  _cert_addr,
        uint8    _cert_target_state,
        bytes32  _latest_valid_oper,
        bytes32  _pre_oper,
        bytes    memory _signature,
        uint256  _blockTime,
        address  _sender
    ) {
        require(opExists[oper_id], "not found");
        DPCEOper storage o = ops[oper_id];
        return (
            o.oper_id,
            o.oper_type,
            o.ca_pk,
            o.cert_addr,
            o.cert_target_state,
            o.latest_valid_oper,
            o.pre_oper,
            o.signature,
            o.blockTime,
            o.sender
        );
    }

    /// @notice 查询某证书的最新操作 oper_id
    function getLatestOper(bytes32 cert_addr) external view returns (bytes32) {
        return latestByCert[cert_addr];
    }

    /// @notice 沿 pre_oper 回溯 limit 条操作，便于审计回放
    function walkOperChain(bytes32 cert_addr, uint256 limit) external view returns (bytes32[] memory) {
        bytes32 cur = latestByCert[cert_addr];
        if (cur == bytes32(0) || limit == 0) {
            return bytes32[0];
        }
        bytes32[] memory out = new bytes32[](limit);
        uint256 i = 0;
        while (cur != bytes32(0) && i < limit) {
            out[i] = cur;
            cur = ops[cur].pre_oper;
            i++;
        }
        // 裁剪长度
        bytes32[] memory trimmed = new bytes32[](i);
        for (uint256 j = 0; j < i; j++) trimmed[j] = out[j];
        return trimmed;
    }

    /// @notice 判断 oper 是否存在（工具函数）
    function exists(bytes32 oper_id) external view returns (bool) {
        return opExists[oper_id];
    }
}
