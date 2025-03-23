import React, { useState, useEffect } from "react";

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

const SpeechRecognitionComponent: React.FC<{
  onRecognized: (transcript: string) => void;
}> = ({ onRecognized }) => {
  const [recognition, setRecognition] = useState<any>(null);

  useEffect(() => {
    const SpeechRecognition =
      window.SpeechRecognition || window.webkitSpeechRecognition;
    if (SpeechRecognition) {
      const recognitionInstance = new SpeechRecognition();
      recognitionInstance.continuous = true;
      recognitionInstance.interimResults = true;
      setRecognition(recognitionInstance);
    }
  }, []);

  useEffect(() => {
    if (recognition) {
      recognition.onresult = (event: SpeechRecognitionEvent) => {
        const transcript = event.results[event.resultIndex][0].transcript;
        onRecognized(transcript);
      };

      recognition.start();

      return () => {
        recognition.stop();
      };
    }
  }, [recognition, onRecognized]);

  return null;
};

export default SpeechRecognitionComponent;
