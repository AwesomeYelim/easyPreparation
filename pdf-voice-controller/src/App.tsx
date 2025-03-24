import React, { useState, useEffect } from "react";
import SpeechRecognitionComponent from "./SpeechRecognition";
import PdfViewer from "./PdfViewer";

const App: React.FC = () => {
  const [currentPage, setCurrentPage] = useState<number>(1);
  const [pdfFile, setPdfFile] = useState<File | null>(null);

  const handleVoiceCommand = (transcript: string) => {
    console.log("Recognized Transcript: ", transcript);

    if (transcript.toLowerCase().includes("next")) {
      setCurrentPage((prevPage) => prevPage + 1);
    }
    if (transcript.toLowerCase().includes("previous")) {
      setCurrentPage((prevPage) => Math.max(prevPage - 1, 1));
    }
  };

  useEffect(() => {
    const timer = setInterval(() => {
      setCurrentPage((prev) => prev + 1);
    }, 5000);
    return () => clearInterval(timer);
  }, []);

  return (
    <div>
      <h1>Voice Controlled PDF Viewer</h1>
      <PdfViewer currentPage={currentPage} pdfFile={pdfFile} />
      <SpeechRecognitionComponent onRecognized={handleVoiceCommand} />
    </div>
  );
};

export default App;
