# frozen_string_literal: true

describe "Top page" do
  context "when a user is not signed in" do
    let!(:work) { create(:work, :with_current_season) }

    before do
      visit "/"
    end

    skip "displays the hero words", js: true do
      expect(page).to have_content("The platform for anime addicts.")
    end
  end
end
