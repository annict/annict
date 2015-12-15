# == Schema Information
#
# Table name: organizations
#
#  id            :integer          not null, primary key
#  name          :string           not null
#  url           :string
#  wikipedia_url :string
#  created_at    :datetime         not null
#  updated_at    :datetime         not null
#
# Indexes
#
#  index_organizations_on_name  (name) UNIQUE
#

class Organization < ActiveRecord::Base
end
