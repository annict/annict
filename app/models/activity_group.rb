# frozen_string_literal: true

# == Schema Information
#
# Table name: activity_groups
#
#  id               :bigint           not null, primary key
#  activities_count :integer          default(0), not null
#  resource_type    :string           not null
#  single           :boolean          default(FALSE), not null
#  created_at       :datetime         not null
#  updated_at       :datetime         not null
#  user_id          :bigint           not null
#
# Indexes
#
#  index_activity_groups_on_created_at  (created_at)
#  index_activity_groups_on_user_id     (user_id)
#
# Foreign Keys
#
#  fk_rails_...  (user_id => users.id)
#
class ActivityGroup < ApplicationRecord
  extend Enumerize

  enumerize :resource_type, in: %w(
    status
    episode_record
    work_record
  ), scope: true

  belongs_to :user
  has_many :activities, dependent: :destroy
end
