# frozen_string_literal: true
# == Schema Information
#
# Table name: reports
#
#  id                 :integer          not null, primary key
#  user_id            :integer          not null
#  root_resource_type :string           not null
#  root_resource_id   :integer          not null
#  resource_type      :string
#  resource_id        :integer
#  created_at         :datetime         not null
#  updated_at         :datetime         not null
#
# Indexes
#
#  index_reports_on_resource_id_and_resource_type            (resource_id,resource_type)
#  index_reports_on_root_resource_id_and_root_resource_type  (root_resource_id,root_resource_type)
#  index_reports_on_user_id                                  (user_id)
#

class Report < ApplicationRecord
  belongs_to :resource, polymorphic: true
  belongs_to :root_resource, polymorphic: true
  belongs_to :user

  validates :user, presence: true
  validates :root_resource, presence: true
end
