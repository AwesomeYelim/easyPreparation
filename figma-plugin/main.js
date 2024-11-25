figma.showUI(__html__, { width: 240, height: 180 });

figma.ui.onmessage = async (msg) => {
    if (msg.type === 'get-selected-node') {
        const selection = figma.currentPage.selection;

        if (selection.length > 0) {
            const nodeInfo = selection.map(node => {
                return {
                    name: node.name,
                    type: node.type,
                    x: node.x,
                    y: node.y,
                    width: node.width,
                    height: node.height
                };
            });
            figma.ui.postMessage({ type: 'node-info', data: nodeInfo });
        } else {
            figma.ui.postMessage({ type: 'node-info', data: 'No node selected' });
        }
    }

    // 예시로 새로운 노드 생성
    if (msg.type === 'create-rectangle') {
        const rect = figma.createRectangle();
        rect.x = 100;
        rect.y = 100;
        rect.resize(100, 100);
        rect.fills = [{ type: 'SOLID', color: { r: 0, g: 0, b: 1 } }];
        figma.currentPage.appendChild(rect);
        figma.ui.postMessage({ type: 'rectangle-created', message: 'Rectangle created!' });
    }
};
