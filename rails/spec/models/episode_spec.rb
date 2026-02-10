# typed: false
# frozen_string_literal: true

RSpec.describe Episode, type: :model do
  describe "#published?" do
    it "status が published のとき true を返すこと" do
      episode_record = FactoryBot.create(:episode, status: :published)

      expect(episode_record.published?).to be true
    end

    it "status が archived のとき false を返すこと" do
      episode_record = FactoryBot.create(:episode, status: :archived)

      expect(episode_record.published?).to be false
    end

    it "status が deleted のとき false を返すこと" do
      episode_record = FactoryBot.create(:episode, status: :deleted)

      expect(episode_record.published?).to be false
    end
  end

  describe "#archived?" do
    it "status が archived のとき true を返すこと" do
      episode_record = FactoryBot.create(:episode, status: :archived)

      expect(episode_record.archived?).to be true
    end

    it "status が published のとき false を返すこと" do
      episode_record = FactoryBot.create(:episode, status: :published)

      expect(episode_record.archived?).to be false
    end

    it "status が deleted のとき false を返すこと" do
      episode_record = FactoryBot.create(:episode, status: :deleted)

      expect(episode_record.archived?).to be false
    end
  end
end
