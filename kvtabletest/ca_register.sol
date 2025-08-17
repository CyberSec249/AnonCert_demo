pragma solidity >=0.6.10 <0.8.20;

contract ca_register {
    enum CAStatus {NotExist, Normal, OnHold, Revoked}

    enum OnHoldReason {
        NotOnHold,
        unspecified,
        eePkCertPrivKeyCompromised,
        caPrivKeyCompromised
    }

    enum RevokeReason {
        NotRevoke,
        unspecified,
        eePkCertPrivKeyCompromised,
        caPrivKeyCompromised,
        superseded,
        cessationOfOperation,
        privilegeWithdrawn,
        weakAlgorithmOrKey
    }
    struct CAInfo {
        uint256 timestamp;
        string certificate;
        CAStatus status;
        OnHoldReason onHoldReason;
        RevokeReason revokeReason;
    }

    mapping(address => CAInfo) private caInfoMap;

    event CARegisterEvent(address indexed caAddress, uint256 timeStamp);
    event CAUpdateEvent(address indexed caAddress, uint256 timeStamp);
    event CAStatusChangeEvent(address indexed caAddress, CAStatus oldStatus, CAStatus newStatus, uint256 timeStamp);

    modifier onlyCA(){
        require(caInfoMap[msg.sender].status == CAStatus.Normal, "permission denied, only registered CA can call this contract");
        _;
    }

    modifier caExists(address caAddress){
        require(caInfoMap[caAddress].status != CAStatus.NotExist, "CA does not exist");
        _;
    }

    modifier caValid(address caAddress){
        require(caInfoMap[caAddress].status == CAStatus.Normal, "CA is not valid");
        _;
    }

    function CARegister(string memory certificate)
        public returns(bool){
        require(caInfoMap[msg.sender].status == CAStatus.NotExist, "CA has been exist");
        require(bytes(certificate).length > 0, "CA Certificate cannot be empty");
        CAInfo memory newCA = CAInfo({
            timestamp: block.timestamp,
            certificate: certificate,
            status: CAStatus.Normal,
            onHoldReason: OnHoldReason.NotOnHold,
            revokeReason: RevokeReason.NotRevoke
        });
        caInfoMap[msg.sender] = newCA;
        emit CARegisterEvent(msg.sender, block.timestamp);
        return true;
    }

    function getCAInfo(address caAddress)
        public view caExists(caAddress) returns(
            uint256 timestamp,
            string memory certificate,
            CAStatus status,
            OnHoldReason onHoldReason,
        RevokeReason revokeReason
    ){
        CAInfo memory ca = caInfoMap[caAddress];
        return(
            ca.timestamp,
            ca.certificate,
            ca.status,
            ca.onHoldReason,
            ca.revokeReason
        );
    }

    function CAUpdate(string memory certificate)
        public onlyCA returns (bool){
        require(bytes(certificate).length > 0, "CA Certificate cannot be empty");

        CAInfo storage ca = caInfoMap[msg.sender];
        ca.certificate = certificate;
        ca.timestamp = block.timestamp;

        emit CAUpdateEvent(msg.sender, block.timestamp);
        return true;
    }

    function OnHoldCA(OnHoldReason reason)
        public onlyCA returns(bool){
            CAInfo storage ca = caInfoMap[msg.sender];
            CAStatus oldStatus = ca.status;
            ca.status = CAStatus.OnHold;
            ca.onHoldReason = reason;
            ca.timestamp = block.timestamp;
            emit CAStatusChangeEvent(msg.sender, oldStatus, CAStatus.OnHold, block.timestamp);
            return true;
    }

    function resumeCA()
    public onlyCA returns(bool){
        require(caInfoMap[msg.sender].status == CAStatus.OnHold, "CA must be on hold to resume");
        CAInfo storage ca = caInfoMap[msg.sender];
        CAStatus oldStatus = ca.status;
        ca.status = CAStatus.Normal;
        ca.onHoldReason = OnHoldReason.NotOnHold;
        ca.timestamp = block.timestamp;
        emit CAStatusChangeEvent(msg.sender, oldStatus, CAStatus.Normal, block.timestamp);
        return true;
    }

    function revokeCA(RevokeReason reason)
        public onlyCA returns (bool){
        CAInfo storage ca = caInfoMap[msg.sender];
        CAStatus oldStatus = ca.status;
        ca.status = CAStatus.Revoked;
        ca.revokeReason = reason;
        ca.timestamp = block.timestamp;
        emit CAStatusChangeEvent(msg.sender, oldStatus, CAStatus.Revoked, block.timestamp);
        return true;
    }
}
