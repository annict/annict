# frozen_string_literal: true

describe "Record" do
  let(:record) { create(:checkin) }

  describe "page loading" do
    it "displays user's record", debug: true do
      visit work_episode_checkin_path(record.work, record.episode, record)
      expect(find("body.checkins-show")).to have_content("おもしろかった")
    end
  end
end
