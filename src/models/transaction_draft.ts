import type { TransactionInfoResponse } from './transaction.ts';
import type { TransactionPictureInfoBasicResponse } from './transaction_picture_info.ts';

export interface TransactionDraftCreateRequest {
    readonly type: number;
    readonly categoryId?: string;
    readonly time: number;
    readonly utcOffset: number;
    readonly accountId?: string;
    readonly destinationAccountId?: string;
    readonly amount: number;
    readonly destinationAmount?: number;
    readonly hideAmount?: boolean;
    readonly tagIds?: string[];
    readonly pictureIds?: string[];
    readonly comment?: string;
    readonly source: string;
    readonly excludeFromBudget?: boolean;
}

export interface TransactionDraftModifyRequest {
    readonly source: string;
    readonly type?: number;
    readonly categoryId?: string;
    readonly time?: number;
    readonly utcOffset?: number;
    readonly accountId?: string;
    readonly destinationAccountId?: string;
    readonly amount?: number;
    readonly destinationAmount?: number;
    readonly hideAmount?: boolean;
    readonly tagIds?: string[];
    readonly pictureIds?: string[];
    readonly comment?: string;
    readonly excludeFromBudget?: boolean;
}

export interface TransactionDraftGetBySourceRequest {
    readonly source: string;
}

export interface TransactionDraftDeleteRequest {
    readonly source: string;
}

export interface TransactionDraftConfirmRequest {
    readonly source: string;
    readonly categoryId?: string;
    readonly accountId?: string;
    readonly destinationAccountId?: string;
    readonly excludeFromBudget?: boolean;
}

export interface TransactionDraftInfoResponse {
    readonly id: string;
    readonly type: number;
    readonly categoryId: string;
    readonly accountId: string;
    readonly destinationAccountId: string;
    readonly time: number;
    readonly utcOffset: number;
    readonly amount: number;
    readonly destinationAmount: number;
    readonly hideAmount: boolean;
    readonly tagIds: string[];
    readonly pictureIds: string[];
    readonly pictures?: TransactionPictureInfoBasicResponse[];
    readonly comment: string;
    readonly excludeFromBudget: boolean;
    readonly source: string;
    readonly complete: boolean;
}

export interface TransactionDraftCountResponse {
    readonly totalCount: number;
}

export type TransactionOrDraftStatus = 'draft' | 'confirmed';

export interface TransactionOrDraftInfoResponse {
    readonly status: TransactionOrDraftStatus;
    readonly draft?: TransactionDraftInfoResponse;
    readonly transaction?: TransactionInfoResponse;
}
