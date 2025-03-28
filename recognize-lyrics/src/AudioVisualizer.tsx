import React, { useEffect, useRef, useState } from "react";
import * as THREE from "three";

interface SpeechRecognitionEvent extends Event {
  results: SpeechRecognitionResultList;
  resultIndex: number;
}

const AudioVisualizer: React.FC = () => {
  const mountRef = useRef<HTMLDivElement>(null);
  const analyserRef = useRef<AnalyserNode | null>(null);
  const dataArrayRef = useRef<Uint8Array | null>(null);
  const meshRef = useRef<THREE.Mesh | null>(null);
  const [isAudioReady, setIsAudioReady] = useState(false);
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
        const scale = 1 + volume / 50;
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
  }, []);

  return (
    <div ref={mountRef} style={{ width: "100vw", height: "100vh" }}>
      {!isAudioReady && <p>Loading Audio...</p>}
    </div>
  );
};

export default AudioVisualizer;
