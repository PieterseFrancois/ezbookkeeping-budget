// Content for the budget help page, shared by the mobile popup and the desktop dialog so the
// two never drift. Kept as data rather than markup because Framework7 and Vuetify render it
// with different components.

export type BudgetHelpBlock =
    | { type: 'p'; text: string }
    | { type: 'list'; items: string[] }
    | { type: 'table'; head: string[]; rows: string[][] }
    | { type: 'note'; text: string }
    // Rendered collapsed so it can be skipped at a glance
    | { type: 'example'; title: string; lines: string[] };

export interface BudgetHelpSection {
    title: string;
    blocks: BudgetHelpBlock[];
}

export const BUDGET_HELP_SECTIONS: BudgetHelpSection[] = [
    {
        title: 'The idea',
        blocks: [
            {
                type: 'p',
                text: 'Every budget line is a category with three numbers: Budgeted (your plan), Actual (what happened), and Remaining (the difference). Lines are grouped into four sections by which way money moves.'
            },
            {
                type: 'table',
                head: ['Section', 'Direction', 'Target means'],
                rows: [
                    ['Income', 'money in', 'what you expect to receive'],
                    ['Expenses', 'money out', 'a ceiling — stay under it'],
                    ['Savings', 'money out, set aside', 'a goal — try to reach it'],
                    ['Cards & Debt', 'money out, paydown', 'a goal — try to reach it']
                ]
            },
            {
                type: 'p',
                text: 'Savings and Cards & Debt are money leaving your spendable cash, the same as an expense. The difference is that it improves your position rather than being consumed, so hitting the target is good rather than bad.'
            },
            {
                type: 'p',
                text: 'Actuals are always filled in for you from your transactions — you only ever set the Budgeted amount. Click any Budgeted cell to edit it.'
            }
        ]
    },
    {
        title: 'Setup you must do first',
        blocks: [
            {
                type: 'p',
                text: 'Income and Expenses need no setup — they are tracked automatically from your existing income and expense categories and their subcategories. Every subcategory becomes a budget row on its own.'
            },
            {
                type: 'p',
                text: 'Savings and Cards & Debt are different. They are driven by transfer categories, and those sections stay empty until you create the parent categories yourself. Go to Transaction Categories, switch to Transfer, and add these two top-level categories:'
            },
            {
                type: 'table',
                head: ['Create this transfer parent', 'Fills this section'],
                rows: [
                    ['Savings & Investments', 'Savings'],
                    ['Loan & Debt', 'Cards & Debt']
                ]
            },
            {
                type: 'note',
                text: 'The names must match exactly, including the "&". If a parent is missing or renamed, its section stays empty with no warning.'
            },
            {
                type: 'p',
                text: 'Then add one subcategory per thing you want to budget. Accounts alone never create budget rows — rows always come from categories — so each savings account or debt you want to track needs its own matching subcategory.'
            },
            {
                type: 'example',
                title: 'Example category setup',
                lines: [
                    'Transfer categories',
                    '  Savings & Investments      ← you must create this',
                    '      Emergency Fund',
                    '      Investments',
                    '      Goal Savings',
                    '  Loan & Debt                ← you must create this',
                    '      Repayment',
                    '',
                    'So if you open a new investments account, add a matching',
                    '"Investments" subcategory, then tag transfers to that',
                    'account with it. Without the category there is nothing',
                    'to budget against.'
                ]
            }
        ]
    },
    {
        title: 'Budget cycle',
        blocks: [
            {
                type: 'p',
                text: 'By default a budget month is the calendar month. If you set a Budget Cycle End Day in settings (for example the 25th), the cycle runs from the 26th of the previous month to the 25th of this one.'
            },
            {
                type: 'p',
                text: 'A cycle is named after the month it ends in, so once you pass the end day you are already working in the next cycle. The date range is always shown under the month strip.'
            },
            {
                type: 'example',
                title: 'Example with end day 25',
                lines: [
                    'The "August" cycle covers 26 July → 25 August.',
                    '',
                    'On 20 August you are still in the August cycle.',
                    'On 26 August the page moves on to September,',
                    'which covers 26 August → 25 September.'
                ]
            }
        ]
    },
    {
        title: 'How transfers are sorted',
        blocks: [
            {
                type: 'p',
                text: 'A transfer is placed by the kind of accounts it moves between, not by its category. Only these three directions are budgeted:'
            },
            {
                type: 'table',
                head: ['From', 'To', 'Section'],
                rows: [
                    ['spendable cash', 'savings / investment', 'Savings'],
                    ['spendable cash', 'credit card / debt', 'Cards & Debt'],
                    ['savings / investment', 'spendable cash', 'Income (a withdrawal)']
                ]
            },
            {
                type: 'p',
                text: 'Everything else is ignored by the budget: moving cash between bank accounts, savings to savings, or drawing new debt. "Spendable cash" means any asset account that is not a savings or investment account.'
            },
            {
                type: 'note',
                text: 'The category decides which row an amount appears on; the accounts decide which section it counts in. If you tag a transfer with a category from the wrong section, the amount can go missing from the page.'
            }
        ]
    },
    {
        title: 'Credit cards and debt',
        blocks: [
            {
                type: 'p',
                text: 'A purchase on a card counts in its own expense category on the day you buy it, exactly like a cash purchase. That keeps your spending categorised.'
            },
            {
                type: 'p',
                text: 'Repaying the card is not counted as spending again. It appears in the Cards & Debt section as progress toward your paydown target, so nothing is double counted.'
            },
            {
                type: 'p',
                text: 'The Cards & Debt panel at the bottom shows, per card: what you charged this cycle, what you paid, and what is still owed. If you charge more than you pay, the owed amount grows — that is the number to watch.'
            },
            {
                type: 'example',
                title: 'Example: a store card month',
                lines: [
                    'You buy R500 groceries and R300 clothing on the card,',
                    'then repay R600 from your bank.',
                    '',
                    'Expenses      Groceries  R500',
                    '              Clothing   R300',
                    'Cards & Debt  Repayment  R600',
                    '',
                    'Panel  charged R800 · paid R600 · owed R200',
                    '',
                    'The R800 is counted once, as spending. The R600 repayment',
                    'is not spending again. Because you charged R800 but only',
                    'paid R600, the debt grew by R200 this cycle.'
                ]
            }
        ]
    },
    {
        title: 'Saving and withdrawing',
        blocks: [
            {
                type: 'p',
                text: 'Money moved into savings shows in the Savings section as progress toward your goal. Withdrawals are not subtracted from it — putting R2000 in and later taking R500 out still shows R2000 saved, because you did save that much.'
            },
            {
                type: 'p',
                text: 'The R500 comes back as a line in the Income section instead, since it is money returning to your spendable cash.'
            },
            {
                type: 'p',
                text: 'You can budget both directions on the same category: a contribution target in Savings and an expected withdrawal target in Income. To add a withdrawal line before any withdrawal has happened, use Add Category and pick the entry marked "Withdrawal".'
            },
            {
                type: 'example',
                title: 'Example: saving then dipping in',
                lines: [
                    'You move R2000 into the emergency fund, then later',
                    'take R500 back out.',
                    '',
                    'Savings   Emergency Fund              R2000',
                    'Income    Emergency Fund (withdrawal)  R500',
                    '',
                    'Savings still reads R2000, not R1500, because you did',
                    'set aside R2000. The R500 shows separately as money',
                    'coming back in. Your account balance is R1500 either way.'
                ]
            }
        ]
    },
    {
        title: 'Leaving a transaction out',
        blocks: [
            {
                type: 'p',
                text: 'A transaction can be excluded from the budget while staying in your records. Useful when you pay for someone and they reimburse you — you still want the transaction, but it should not count against your budget.'
            },
            {
                type: 'p',
                text: 'Open the transaction and turn off its budget inclusion. Exclusions apply to every section, not just expenses.'
            }
        ]
    },
    {
        title: 'Tidying the page',
        blocks: [
            {
                type: 'p',
                text: 'Rows you do not use can be hidden with the eye icon. A row can only be hidden while it has no budget and no activity, so nothing with real numbers can disappear.'
            },
            {
                type: 'p',
                text: 'Use Add Category to bring hidden rows back. Copy Budget takes the targets from another month so you do not start each cycle from scratch.'
            }
        ]
    }
];
