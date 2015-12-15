# == Schema Information
#
# Table name: member_participations
#
#  id              :integer          not null, primary key
#  person_id       :integer          not null
#  organization_id :integer          not null
#  created_at      :datetime         not null
#  updated_at      :datetime         not null
#
# Indexes
#
#  index_member_participations_on_organization_id                (organization_id)
#  index_member_participations_on_person_id                      (person_id)
#  index_member_participations_on_person_id_and_organization_id  (person_id,organization_id) UNIQUE
#

class MemberParticipation < ActiveRecord::Base
end
