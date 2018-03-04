# frozen_string_literal: true

describe "Comment" do
  describe "create a comment" do
    let(:record) { create(:record) }
    let(:user) { create(:registered_user) }

    before do
      login_as(user, scope: :user)
    end

    it "creates a comment" do
      visit "/@#{record.user.username}/records/#{record.id}"

      within "#new_comment" do
        fill_in "comment[body]", with: "Hahaha"
      end
      find("button").click

      expect(find(".c-record-comment")).to have_content("Hahaha")
    end
  end
end
