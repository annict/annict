# typed: false
# frozen_string_literal: true

RSpec.describe Staff, type: :model do
  describe ".with_kept_person" do
    it "人物が非公開・削除済みのスタッフを除き、団体のスタッフは残すこと" do
      published_person_staff = FactoryBot.create(:staff, resource: FactoryBot.create(:person))
      unpublished_person_staff = FactoryBot.create(:staff, resource: FactoryBot.create(:person, :unpublished))
      deleted_person_staff = FactoryBot.create(:staff, resource: FactoryBot.create(:person, :deleted))
      organization_staff = FactoryBot.create(:staff, resource: FactoryBot.create(:organization))

      staffs = Staff.with_kept_person

      expect(staffs).to include(published_person_staff, organization_staff)
      expect(staffs).not_to include(unpublished_person_staff, deleted_person_staff)
    end

    it "既存の絞り込みと組み合わせたとき、既存の条件も維持されること" do
      work = FactoryBot.create(:work)
      kept_staff = FactoryBot.create(:staff, work:)
      unpublished_staff = FactoryBot.create(:staff, :unpublished, work:)
      other_work_staff = FactoryBot.create(:staff)

      staffs = work.staffs.only_kept.with_kept_person

      expect(staffs).to contain_exactly(kept_staff)
      expect(staffs).not_to include(unpublished_staff, other_work_staff)
    end
  end
end
