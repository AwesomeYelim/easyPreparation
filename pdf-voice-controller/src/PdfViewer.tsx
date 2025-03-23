import React, { useEffect, useRef, useState } from "react";
import { getDocument } from "pdfjs-dist";

interface PdfViewerProps {
  pdfUrl: string;
  currentPage: number;
}

const PdfViewer: React.FC<PdfViewerProps> = ({ pdfUrl, currentPage }) => {
  const [pdf, setPdf] = useState<any>(null);
  const canvasRef = useRef<HTMLCanvasElement>(null);

  //   useEffect(() => {
  //     const fetchPdf = async () => {
  //       try {
  //         const pdfDoc = await getDocument(pdfUrl).promise;
  //         setPdf(pdfDoc);
  //       } catch (err) {
  //         console.error("PDF 로딩 중 오류: ", err);
  //       }
  //     };

  //     fetchPdf();
  //   }, [pdfUrl]);

  useEffect(() => {
    if (pdf) {
      renderPage(currentPage);
    }
  }, [pdf, currentPage]);

  const renderPage = (pageNum: number) => {
    if (pdf && canvasRef.current) {
      const canvas = canvasRef.current;
      const context = canvas.getContext("2d");
      const viewport = pdf.getPage(pageNum).getViewport({ scale: 1.5 });
      if (context) {
        canvas.height = viewport.height;
        canvas.width = viewport.width;
        pdf.getPage(pageNum).then((page: any) => {
          page.render({ canvasContext: context, viewport });
        });
      }
    }
  };

  return <canvas ref={canvasRef} />;
};

export default PdfViewer;
