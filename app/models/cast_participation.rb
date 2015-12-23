# == Schema Information
#
# Table name: cast_participations
#
#  id         :integer          not null, primary key
#  person_id  :integer          not null
#  work_id    :integer          not null
#  name       :string
#  part       :string           not null
#  created_at :datetime         not null
#  updated_at :datetime         not null
#
# Indexes
#
#  index_cast_participations_on_person_id  (person_id)
#  index_cast_participations_on_work_id    (work_id)
#

class CastParticipation < ActiveRecord::Base
  include DbActivityMethods

  belongs_to :person
  belongs_to :work

  validates :person_id, presence: true
  validates :work_id, presence: true
  validates :part, presence: true
end
