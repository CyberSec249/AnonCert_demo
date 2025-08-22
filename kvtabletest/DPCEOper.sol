pragma solidity>=0.4.24 <0.6.11;

import "./Table.sol";

contract DPCEOper {
    event SetResult(int256 count);

    KVTableFactory tableFactory;
    string constant TABLE_NAME = "t_certOper";

    constructor() public {
        //The fixed address is 0x1010 for KVTableFactory
        tableFactory = KVTableFactory(0x1010);
        // the parameters of createTable are tableName,keyField,"vlaueFiled1,vlaueFiled2,vlaueFiled3,..."
        tableFactory.createTable(TABLE_NAME, "oper_id", "oper_type, ca_pk, cert_addr, cert_target_state, latest_valid_oper, pre_oper, signature");
    }

    //get record
    function getCertOPer(string memory oper_id) public view returns (
        bool,
        string memory oper_type,
        string memory ca_pk,
        string memory cert_addr,
        string memory cert_target_state,
        string memory latest_valid_oper,
        string memory pre_oper,
        string memory signature
    ) {
        KVTable table = tableFactory.openTable(TABLE_NAME);
        bool ok = false;
        Entry entry;
        (ok, entry) = table.get(oper_id);

        if (ok) {
            oper_type = entry.getString("oper_type");
            ca_pk = entry.getString("ca_pk");
            cert_addr = entry.getString("cert_addr");
            cert_target_state = entry.getString("cert_target_state");
            latest_valid_oper = entry.getString("latest_valid_oper");
            pre_oper = entry.getString("pre_oper");
            signature = entry.getString("signature");
        }
        return;
    }

    //set record
    function setCertOPer(string memory oper_id, string memory oper_type, string memory ca_pk, string memory cert_addr, string memory cert_target_state, string memory latest_valid_oper, string memory pre_oper, string memory signature)
    public
    returns (int256)
    {
        KVTable table = tableFactory.openTable(TABLE_NAME);
        Entry entry = table.newEntry();
        // the length of entry's field value should < 16MB
        entry.set("oper_id", oper_id);
        entry.set("oper_type", oper_type);
        entry.set("ca_pk", ca_pk);
        entry.set("cert_addr", cert_addr);
        entry.set("cert_target_state", cert_target_state);
        entry.set("latest_valid_oper", latest_valid_oper);
        entry.set("pre_oper", pre_oper);
        entry.set("signature", signature);

        // the first parameter length of set should <= 255B
        int256 count = table.set(oper_id, entry);
        emit SetResult(count);
        return count;
    }
}