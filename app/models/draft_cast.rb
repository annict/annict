# == Schema Information
#
# Table name: draft_casts
#
#  id         :integer          not null, primary key
#  cast_id    :integer
#  person_id  :integer          not null
#  work_id    :integer          not null
#  name       :string           not null
#  part       :string           not null
#  created_at :datetime         not null
#  updated_at :datetime         not null
#
# Indexes
#
#  index_draft_casts_on_cast_id    (cast_id)
#  index_draft_casts_on_person_id  (person_id)
#  index_draft_casts_on_work_id    (work_id)
#

class DraftCast < ActiveRecord::Base
  include DraftCommon
  include CastCommon

  belongs_to :origin, class_name: "Cast", foreign_key: :cast_id
end
