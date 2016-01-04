# == Schema Information
#
# Table name: draft_staffs
#
#  id          :integer          not null, primary key
#  staff_id    :integer
#  person_id   :integer          not null
#  work_id     :integer          not null
#  name        :string           not null
#  role        :string           not null
#  role_other  :string
#  sort_number :integer          default(0), not null
#  created_at  :datetime         not null
#  updated_at  :datetime         not null
#
# Indexes
#
#  index_draft_staffs_on_person_id    (person_id)
#  index_draft_staffs_on_sort_number  (sort_number)
#  index_draft_staffs_on_staff_id     (staff_id)
#  index_draft_staffs_on_work_id      (work_id)
#

class DraftStaff < ActiveRecord::Base
  include DraftCommon
  include StaffCommon

  belongs_to :origin, class_name: "Staff", foreign_key: :staff_id
end
