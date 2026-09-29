import { ref } from 'vue';
import { defineStore } from 'pinia';

import type { TransactionDraftInfoResponse } from '@/models/transaction_draft.ts';

import services from '@/lib/services.ts';

export const useTransactionDraftsStore = defineStore('transactionDrafts', () => {
    const draftCount = ref<number>(0);

    // holds the draft being handed off to the mobile add/edit page (a full routed page, not a
    // modal like the desktop dialog), set right before navigation and consumed once on that
    // page's init(); avoids re-fetching or encoding the whole draft (incl. pictures) into the URL
    const pendingEditDraft = ref<TransactionDraftInfoResponse | null>(null);

    function refreshDraftCount(): Promise<number> {
        return services.getTransactionDraftCount().then(response => {
            if (response.data && response.data.success && response.data.result) {
                draftCount.value = response.data.result.totalCount;
            }

            return draftCount.value;
        }).catch(() => {
            return draftCount.value;
        });
    }

    function setPendingEditDraft(draft: TransactionDraftInfoResponse): void {
        pendingEditDraft.value = draft;
    }

    function takePendingEditDraft(source: string): TransactionDraftInfoResponse | null {
        const draft = pendingEditDraft.value;
        pendingEditDraft.value = null;

        if (draft && draft.source === source) {
            return draft;
        }

        return null;
    }

    return {
        draftCount,
        refreshDraftCount,
        setPendingEditDraft,
        takePendingEditDraft
    };
});
