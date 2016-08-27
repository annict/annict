# frozen_string_literal: true

describe "Record" do
  describe "detail page" do
    let(:record) { create(:checkin) }

    it "displays user's record" do
      visit work_episode_checkin_path(record.work, record.episode, record)
      expect(find("body.checkins-show")).to have_content("おもしろかった")
    end
  end

  describe "edit page" do
    let(:user) { create(:registered_user) }
    let(:record) { create(:checkin, user: user) }

    before do
      login_as(user, scope: :user)
    end

    it "can update user's record", js: true do
      visit edit_work_episode_checkin_path(record.work, record.episode, record)
      within(".edit_checkin") do
        fill_in "checkin[comment]", with: "So cool!"
        click_button "更新する"
      end
      expect(find("body.checkins-show")).to have_content("So cool!")
    end
  end
end
