import React from 'react';
import './Modal.css'
interface ModalProps {
    isOpen: boolean;
    onClose: () => void;
    children: React.JSX.Element;
}

const Modal = (props: ModalProps) => {
    if (!props.isOpen) return null;

    return (

        <div className={"modal"}>
            <span className="close" onClick={props.onClose}>&times;</span>
            <div className={"modal-content"}>
                {props.children}
            </div>
        </div>

    );
};

export function ShowModal(content: React.JSX.Element): React.JSX.Element {
    const [isOpen, setIsOpen] = React.useState(true);
    return (
        <div>
            <Modal isOpen={isOpen}
                   onClose={() => setIsOpen(false)}
                   children={content}
            >
            </Modal>
        </div>
    );
}

export default Modal;
