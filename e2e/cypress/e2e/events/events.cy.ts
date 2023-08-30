describe('events', () => {
  beforeEach(() => {
    cy.context().as('ctx');
  });

  it('events can be filtered', () => {
    const eventTypeEnglish = 'Instance added';
    cy.visit('/events');
    cy.get('[data-e2e="event-type-cell"]').should('have.length', 20);
    cy.get('[data-e2e="open-filter-button"]').click();
    cy.get('[data-e2e="event-type-filter-checkbox"]').click();
    cy.get('#mat-select-value-1').click();
    cy.contains('mat-option', eventTypeEnglish).click();
    cy.get('body').click();
    cy.get('[data-e2e="filter-finish-button"]').click();
    cy.contains('[data-e2e="event-type-cell"]', eventTypeEnglish).should('have.length.at.least', 1);
  });
});
