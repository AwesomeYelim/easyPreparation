import React, { useEffect, useState } from "react";
import { Document, Page } from "react-pdf";

interface PdfViewerProps {
  currentPage: number;
  pdfFile: File | null;
}

const PdfViewer: React.FC<PdfViewerProps> = ({ currentPage, pdfFile }) => {
  const [numPages, setNumPages] = useState<number | null>(null);
  const [scale, setScale] = useState<number>(1);

  const onLoadSuccess = ({ numPages }: any) => {
    setNumPages(numPages);
  };

  useEffect(() => {
    const updateScale = () => {
      const pageWidth = window.innerWidth;
      const pageHeight = window.innerHeight;

      const scaleFactor = Math.min(pageWidth / 1000, pageHeight / 1400);
      setScale(scaleFactor);
    };

    window.addEventListener("resize", updateScale);

    updateScale();

    return () => {
      window.removeEventListener("resize", updateScale);
    };
  }, []);

  return (
    <div>
      {pdfFile && (
        <Document
          file={URL.createObjectURL(pdfFile)}
          onLoadSuccess={onLoadSuccess}
        >
          <Page pageNumber={currentPage} scale={scale} />
        </Document>
      )}
      <div>
        <button
          onClick={() => setCurrentPage(currentPage - 1)}
          disabled={currentPage <= 1}
        >
          Previous Page
        </button>
        <button
          onClick={() => setCurrentPage(currentPage + 1)}
          disabled={currentPage >= numPages!}
        >
          Next Page
        </button>
      </div>
    </div>
  );
};

export default PdfViewer;
