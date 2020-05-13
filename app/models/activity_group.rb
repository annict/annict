# frozen_string_literal: true

# == Schema Information
#
# Table name: activity_groups
#
#  id               :bigint           not null, primary key
#  action           :string           not null
#  activities_count :integer          default(0), not null
#  single           :boolean          default(FALSE), not null
#  created_at       :datetime         not null
#  updated_at       :datetime         not null
#
# Indexes
#
#  index_activity_groups_on_created_at  (created_at)
#
class ActivityGroup < ApplicationRecord
end
