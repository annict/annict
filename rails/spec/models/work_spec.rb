# typed: false
# frozen_string_literal: true

RSpec.describe Work, type: :model do
  describe "#published?" do
    it "unpublished_at が nil のとき true を返すこと" do
      work_record = FactoryBot.create(:work, unpublished_at: nil)

      expect(work_record.published?).to be true
    end

    it "unpublished_at が設定されているとき false を返すこと" do
      work_record = FactoryBot.create(:work, unpublished_at: Time.zone.now)

      expect(work_record.published?).to be false
    end
  end
end
