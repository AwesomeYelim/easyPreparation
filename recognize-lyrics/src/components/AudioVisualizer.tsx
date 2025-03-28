import React, { useEffect, useRef, useState } from "react";
import * as THREE from "three";

interface SpeechRecognitionEvent extends Event {
  results: SpeechRecognitionResultList;
  resultIndex: number;
}

declare global {
  interface Window {
    SpeechRecognition: any;
    webkitSpeechRecognition: any;
  }
}

const songLyrics = [
  "의지했던 모든 것 변해가고\n억울한 마음은 커져가네",
  "부끄럼 없이 살고 싶은 맘\n주님 아시네",
  "모든 일을 선으로 이겨내고\n죄의 유혹을 따르지 않네",
  "나를 구원하신 영원한 그 사랑\n크신 그 은혜 날 붙드시네",
  "주어진 내 삶이 작게만 보여도\n선하신 주 나를 이끄심 보네",
  "중심을 보시는 주님만 따르네\n날 택하신 주만 의지해",
  "보이는 상황에 무너질지라도\n예수 능력이 나를 붙드네",
  "보이지 않아도 주님만 따르네\n내 평생 주님을 노래하리라",
  "모든 일을 선으로 이겨내고\n죄의 유혹을 따르지 않네",
  "나를 구원하신 영원한 그 사랑\n크신 그 은혜 날 붙드시네",
  "주어진 내 삶이 작게만 보여도\n선하신 주 나를 이끄심 보네",
  "중심을 보시는 주님만 따르네\n날 택하신 주만 의지해",
  "보이는 상황에 무너질지라도\n예수 능력이 나를 붙드네",
  "보이지 않아도 주님만 따르네\n내 평생 주님을 노래하리라",
];

const AudioVisualizer: React.FC = () => {
  const mountRef = useRef<HTMLDivElement>(null);
  const analyserRef = useRef<AnalyserNode | null>(null);
  const dataArrayRef = useRef<Uint8Array | null>(null);
  const meshRef = useRef<THREE.Mesh | null>(null);
  const [isAudioReady, setIsAudioReady] = useState(false);
  const [lyrics, setLyrics] = useState<string[]>([]);
  const [currentLine, setCurrentLine] = useState<number>(0);
  let renderer: THREE.WebGLRenderer | null = null;

  useEffect(() => {
    if (!mountRef.current) return;
    console.log("Initializing Three.js...");

    // Three.js 기본 설정
    const scene = new THREE.Scene();
    const camera = new THREE.PerspectiveCamera(
      75,
      window.innerWidth / window.innerHeight,
      0.1,
      1000
    );
    camera.position.z = 5;

    // 기존 renderer가 있으면 제거하고 새로 생성
    if (renderer) {
      renderer.dispose(); // 이전 renderer 리소스 해제
      mountRef.current.removeChild(renderer.domElement); // 이전 캔버스 제거
    }

    renderer = new THREE.WebGLRenderer();
    renderer.setSize(window.innerWidth, window.innerHeight);
    mountRef.current.appendChild(renderer.domElement); // 새로운 캔버스 추가

    // 기본 조명 추가
    scene.add(new THREE.AmbientLight(0xffffff, 0.5)); // 주변광 추가
    const light = new THREE.PointLight(0xffffff, 2, 100);
    light.position.set(5, 5, 5);
    scene.add(light);

    const geometry = new THREE.SphereGeometry(1, 64, 64);
    const material = new THREE.MeshStandardMaterial({
      color: 0xffffff, // 흰색
      transparent: true,
      opacity: 0.5,
      wireframe: false,
    });
    const mesh = new THREE.Mesh(geometry, material);
    scene.add(mesh);
    meshRef.current = mesh;

    // 마이크 오디오 설정
    const setupAudio = async () => {
      try {
        const stream = await navigator.mediaDevices.getUserMedia({
          audio: true,
        });
        const audioContext = new AudioContext();
        const source = audioContext.createMediaStreamSource(stream);
        const analyser = audioContext.createAnalyser();
        analyser.fftSize = 512;
        source.connect(analyser);

        analyserRef.current = analyser;
        dataArrayRef.current = new Uint8Array(analyser.frequencyBinCount);

        setIsAudioReady(true); // 오디오 설정 완료 후 상태 변경

        console.log("🔊 오디오 분석기 설정 완료!");
      } catch (error) {
        console.error("오디오 입력을 가져오는 데 실패했습니다:", error);
      }
    };

    setupAudio();

    // 애니메이션 루프
    const animate = () => {
      requestAnimationFrame(animate);

      if (renderer) {
        renderer.render(scene, camera);
      }

      if (analyserRef.current && dataArrayRef.current && meshRef.current) {
        analyserRef.current.getByteFrequencyData(dataArrayRef.current);

        // 소리 크기 분석 → 구체 크기 조절
        const volume =
          dataArrayRef.current.reduce((a, b) => a + b, 0) /
          dataArrayRef.current.length;
        const scale = 1 + volume / 100;
        meshRef.current.scale.set(scale, scale, scale);

        const bass =
          dataArrayRef.current.slice(0, 32).reduce((a, b) => a + b, 0) / 32;
        const treble =
          dataArrayRef.current.slice(100, 256).reduce((a, b) => a + b, 0) / 156;

        // 색상 업데이트 (bass와 treble 값에 따른 색상 변화)
        const color = new THREE.Color(
          Math.min(bass / 256, 1),
          Math.min(1 - bass / 256, 1),
          Math.min(treble / 256, 1)
        );
        (meshRef.current.material as THREE.MeshStandardMaterial).color.set(
          color
        );

        // 물리적 효과: 진동, 변형 등 추가 가능
        const time = performance.now() * 0.002;
        meshRef.current.rotation.x = time;
        meshRef.current.rotation.y = time;
      }
    };

    animate();

    // 창 크기 조절 대응
    const onResize = () => {
      if (renderer && camera) {
        camera.aspect = window.innerWidth / window.innerHeight;
        camera.updateProjectionMatrix();
        renderer.setSize(window.innerWidth, window.innerHeight);
      }
    };
    window.addEventListener("resize", onResize);

    // 음성 인식 설정 (SpeechRecognition 또는 webkitSpeechRecognition 사용)
    const recognition =
      window.SpeechRecognition || window.webkitSpeechRecognition;
    if (!recognition) {
      console.error("SpeechRecognition API is not supported in this browser.");
      return;
    }

    const recognitionInstance = new recognition();
    recognitionInstance.lang = "ko-KR"; // 한국어로 설정
    recognitionInstance.continuous = true;
    recognitionInstance.interimResults = true; // 중간 결과 허용

    recognitionInstance.onresult = (event: SpeechRecognitionEvent) => {
      const transcript = event.results[event.resultIndex][0].transcript;
      console.log("Recognized speech:", transcript);

      // transcript와 매칭되는 가사 줄을 찾기
      const matchedLine = songLyrics.findIndex((line) =>
        line.includes(transcript)
      );

      if (matchedLine !== -1) {
        // 해당 줄을 찾았으면, 그 줄로 이동
        setLyrics([songLyrics[matchedLine]]);
        setCurrentLine(matchedLine); // 매칭된 줄로 currentLine 설정
      }
    };

    recognitionInstance.start(); // 음성 인식 시작

    return () => {
      console.log("Cleaning up...");
      window.removeEventListener("resize", onResize);
      if (renderer) {
        renderer.dispose();
      }
      scene.clear();
      if (mountRef.current && renderer) {
        mountRef.current.removeChild(renderer.domElement);
      }
    };
  }, [currentLine]);

  return (
    <div ref={mountRef} style={{ width: "100vw", height: "100vh" }}>
      {!isAudioReady && <p>Loading Audio...</p>}
      <div
        style={{
          position: "absolute",
          width: "70%",
          top: "50%",
          left: "50%",
          transform: "translate(-50%, -50%)",
          color: "white",
          fontSize: "5rem",
          fontWeight: "bold",
          fontFamily: "Nanum, sans-serif",
          textAlign: "center",
          lineHeight: "1.5",
        }}
      >
        {lyrics.map((line, index) => (
          <p key={index}>{line}</p>
        ))}
      </div>
    </div>
  );
};

export default AudioVisualizer;
