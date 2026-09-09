# typed: false
# frozen_string_literal: true

RSpec.describe Episode, type: :model do
  describe "#published?" do
    it "unpublished_at が nil のとき true を返すこと" do
      episode_record = FactoryBot.create(:episode, unpublished_at: nil)

      expect(episode_record.published?).to be true
    end

    it "unpublished_at が設定されているとき false を返すこと" do
      episode_record = FactoryBot.create(:episode, unpublished_at: Time.zone.now)

      expect(episode_record.published?).to be false
    end
  end
end
