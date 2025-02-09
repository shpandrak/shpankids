import React from 'react';


class UiCtx {
    constructor(modalSetter: (modal?: React.JSX.Element) => void) {
        this.modalSetter = modalSetter;
    }

    private readonly modalSetter: (modal?: React.JSX.Element) => void;

    public showModal(content: React.JSX.Element): void {
        this.modalSetter(content)
    }

    public hideModal(): void {
        this.modalSetter(undefined)
    }


}

export default UiCtx;
