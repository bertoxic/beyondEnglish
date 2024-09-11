let ws;
let localStream;
let peerConnections = {};
let myPeerID;

const createRoomBtn = document.getElementById('createRoom');
const joinRoomBtn = document.getElementById('joinRoom');
const leaveRoomBtn = document.getElementById('leaveRoom');
const roomInput = document.getElementById('roomInput');
const videoContainer = document.getElementById('videoContainer');

createRoomBtn.addEventListener('click', createRoom);
joinRoomBtn.addEventListener('click', joinRoom);
leaveRoomBtn.addEventListener('click', leaveRoom);

function connectWebSocket() {
   // ws = new WebSocket('ws://localhost:8080/ws');
    ws = new WebSocket('ws://192.168.195.190:8080/ws');
    
    ws.onmessage = (event) => {
        const message = JSON.parse(event.data);
        console.log('Received message:', message);
        handleSignalingMessage(message);
    };

    ws.onopen = () => {
        console.log('WebSocket connected');
    };

    ws.onerror = (error) => {
        console.error('WebSocket error:', error);
    };
}
function handleSignalingMessage(message) {
    console.log('Handling signaling message:', message);
    switch (message.type) {
        case 'peer-id':
            myPeerID = message.data.peerID;
            console.log('Received my peer ID:', myPeerID);
            break;
        case 'user-joined':
           /// console.log('New user joined:', message.data.peerID);
            console.log('New user joined:', message.data);
            initiatePeerConnection(message.data.peerID);
            break;
        case 'offer':
            handleOffer(message.data);
            break;
        case 'answer':
            handleAnswer(message.data);
            break;
        case 'ice-candidate':
            handleICECandidate(message.data);
            break;
    }
}
function createRoom() {
    const roomName = roomInput.value;
    if (!roomName) return;

    setupLocalStream().then(() => {
        ws.send(JSON.stringify({ type: 'create', data: { roomName } }));
        console.log('Created room:', roomName);
    });
}

function joinRoom() {
    const roomName = roomInput.value;
    if (!roomName) return;

    setupLocalStream().then(() => {
        ws.send(JSON.stringify({ type: 'join', data: { roomName } }));
        console.log('Joined room:', roomName);
    });
}

function leaveRoom() {
    const roomName = roomInput.value;
    if (!roomName) return;

    ws.send(JSON.stringify({ type: 'leave', data: { roomName } }));
    cleanupRTC();
    console.log('Left room:', roomName);
}

async function setupLocalStream() {
    try {
        localStream = await navigator.mediaDevices.getUserMedia({ video: true, audio: true });
        addVideoStream('local', localStream);
        console.log('Local stream set up');
    } catch (error) {
        console.error('Error accessing media devices:', error);
    }
}

function addVideoStream(id, stream) {
    const video = document.createElement('video');
    video.srcObject = stream;
    video.id = id;
    video.autoplay = true;
    video.playsInline = true;
    videoContainer.appendChild(video);
    console.log('Added video stream:', id);
}



function initiatePeerConnection(peerID) {
    console.log('Initiating peer connection for:', peerID);
    const pc = createPeerConnection(peerID);
    pc.createOffer().then(offer => {
        return pc.setLocalDescription(offer);
    }).then(() => {
        ws.send(JSON.stringify({
            type: 'offer',
            data: { offer: pc.localDescription, peerID, roomName: roomInput.value }
        }));
    }).catch(error => console.error('Error creating offer:', error));
}


function createPeerConnection(peerID) {
    console.log('Creating peer connection for:', peerID);
    const pc = new RTCPeerConnection({
        iceServers: [{ urls: 'stun:stun.l.google.com:19302' }]
    });
    
    pc.onicecandidate = (event) => {
        if (event.candidate) {
            ws.send(JSON.stringify({
                type: 'ice-candidate',
                data: { candidate: event.candidate, peerID }
            }));
        }
    };

    pc.ontrack = (event) => {
        console.log('Received remote track from:', peerID);
        addVideoStream(peerID, event.streams[0]);
    };

    localStream.getTracks().forEach(track => pc.addTrack(track, localStream));

    peerConnections[peerID] = pc;
    return pc;
}

function handleOffer(data) {
    console.log('Handling offer from:', data.peerID);
    const pc = createPeerConnection(data.peerID);
    pc.setRemoteDescription(new RTCSessionDescription(data.offer))
        .then(() => pc.createAnswer())
        .then(answer => pc.setLocalDescription(answer))
        .then(() => {
            ws.send(JSON.stringify({
                type: 'answer',
                data: { answer: pc.localDescription, peerID: data.peerID, roomName: roomInput.value }
            }));
        })
        .catch(error => console.error('Error handling offer:', error));
}

async function handleAnswer(data) {
    console.log('Handling answer from:', data.peerID);
    const pc = peerConnections[data.peerID];
    if (pc) {
        await pc.setRemoteDescription(new RTCSessionDescription(data.answer));
    } else {
        console.error('No peer connection found for:', data.peerID);
    }
}

function handleICECandidate(data) {
    console.log('Handling ICE candidate for:', data.peerID);
    const pc = peerConnections[data.peerID];
    if (pc) {
        pc.addIceCandidate(new RTCIceCandidate(data.candidate))
            .catch(error => console.error('Error adding ICE candidate:', error));
    } else {
        console.error('No peer connection found for:', data.peerID);
    }
}

function cleanupRTC() {
    localStream.getTracks().forEach(track => track.stop());
    Object.values(peerConnections).forEach(pc => pc.close());
    peerConnections = {};
    videoContainer.innerHTML = '';
    console.log('Cleaned up RTC connections');
}

connectWebSocket();