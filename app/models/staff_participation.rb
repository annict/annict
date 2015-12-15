# == Schema Information
#
# Table name: staff_participations
#
#  id         :integer          not null, primary key
#  person_id  :integer          not null
#  work_id    :integer          not null
#  name       :string
#  role       :string           not null
#  role_other :string
#  created_at :datetime         not null
#  updated_at :datetime         not null
#
# Indexes
#
#  index_staff_participations_on_person_id  (person_id)
#  index_staff_participations_on_work_id    (work_id)
#

class StaffParticipation < ActiveRecord::Base
  extend Enumerize

  enumerize :role, in: %w(
    animation_director
    art_director
    character_design
    chief_animation_director
    chief_director
    director
    music
    original_character_design
    original_creator
    other
    script
    series_composition
    sound_director
  )

  belongs_to :person
  belongs_to :work
end
