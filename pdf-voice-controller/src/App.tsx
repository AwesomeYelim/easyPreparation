import React, { useState } from "react";
import SpeechRecognitionComponent from "./SpeechRecognition";
import PdfViewer from "./PdfViewer";

const App: React.FC = () => {
  const [currentPage, setCurrentPage] = useState<number>(1);
  const pdfUrl = "../주만바라 볼찌라.pdf";

  const handleVoiceCommand = (transcript: string) => {
    console.log("Recognized Transcript: ", transcript);

    if (transcript.toLowerCase().includes("next")) {
      setCurrentPage((prevPage) => prevPage + 1);
    }
    if (transcript.toLowerCase().includes("previous")) {
      setCurrentPage((prevPage) => Math.max(prevPage - 1, 1));
    }
  };

  return (
    <div>
      <h1>Voice Controlled PDF Viewer</h1>
      <PdfViewer pdfUrl={pdfUrl} currentPage={currentPage} />
      <SpeechRecognitionComponent onRecognized={handleVoiceCommand} />
    </div>
  );
};

export default App;
