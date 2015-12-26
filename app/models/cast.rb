# == Schema Information
#
# Table name: casts
#
#  id         :integer          not null, primary key
#  person_id  :integer          not null
#  work_id    :integer          not null
#  name       :string           not null
#  part       :string           not null
#  created_at :datetime         not null
#  updated_at :datetime         not null
#
# Indexes
#
#  index_cast_participations_on_person_id  (person_id)
#  index_cast_participations_on_work_id    (work_id)
#

class Cast < ActiveRecord::Base
  include DbActivityMethods
  include CastCommon
end
