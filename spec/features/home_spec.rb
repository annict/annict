# frozen_string_literal: true

describe "Top page" do
  context "when a user is not signed in" do
    let!(:work) { create(:work, :with_item, :with_current_season) }

    before do
      visit "/"
    end

    it "displays the hero words" do
      expect(page).to have_content("The platform for Anime addicts.")
    end
  end
end
