# frozen_string_literal: true

describe "Comment" do
  describe "create a comment" do
    let(:episode_record) { create(:episode_record) }
    let(:user) { create(:registered_user) }

    before do
      login_as(user, scope: :user)
    end

    it "creates a comment" do
      visit "/@#{episode_record.user.username}/records/#{episode_record.record.id}"

      within "#new_comment" do
        fill_in "comment[body]", with: "Hahaha"

        scroll_to("button")
        find("button").click
      end

      expect(find(".c-record-comment")).to have_content("Hahaha")
    end
  end
end
