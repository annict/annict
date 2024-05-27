# typed: false
# frozen_string_literal: true

# == Schema Information
#
# Table name: userland_categories
#
#  id                      :bigint           not null, primary key
#  name                    :string           not null
#  name_en                 :string           not null
#  sort_number             :integer          default(0), not null
#  userland_projects_count :integer          default(0), not null
#  created_at              :datetime         not null
#  updated_at              :datetime         not null
#

class UserlandCategory < ApplicationRecord
  has_many :userland_projects
end
