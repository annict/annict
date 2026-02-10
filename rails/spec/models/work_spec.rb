# typed: false
# frozen_string_literal: true

RSpec.describe Work, type: :model do
  describe "#published?" do
    it "status が published のとき true を返すこと" do
      work_record = FactoryBot.create(:work, status: :published)

      expect(work_record.published?).to be true
    end

    it "status が archived のとき false を返すこと" do
      work_record = FactoryBot.create(:work, status: :archived)

      expect(work_record.published?).to be false
    end

    it "status が deleted のとき false を返すこと" do
      work_record = FactoryBot.create(:work, status: :deleted)

      expect(work_record.published?).to be false
    end
  end

  describe "#archived?" do
    it "status が archived のとき true を返すこと" do
      work_record = FactoryBot.create(:work, status: :archived)

      expect(work_record.archived?).to be true
    end

    it "status が published のとき false を返すこと" do
      work_record = FactoryBot.create(:work, status: :published)

      expect(work_record.archived?).to be false
    end

    it "status が deleted のとき false を返すこと" do
      work_record = FactoryBot.create(:work, status: :deleted)

      expect(work_record.archived?).to be false
    end
  end
end
