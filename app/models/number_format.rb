# == Schema Information
#
# Table name: number_formats
#
#  id          :integer          not null, primary key
#  name        :string           not null
#  data        :string           default([]), not null, is an Array
#  created_at  :datetime         not null
#  updated_at  :datetime         not null
#  sort_number :integer          default(0), not null
#
# Indexes
#
#  index_number_formats_on_name  (name) UNIQUE
#

class NumberFormat < ActiveRecord::Base
end
