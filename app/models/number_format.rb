# frozen_string_literal: true

# == Schema Information
#
# Table name: number_formats
#
#  id          :bigint           not null, primary key
#  data        :string           default([]), not null, is an Array
#  format      :string           default(""), not null
#  name        :string           not null
#  sort_number :integer          default(0), not null
#  created_at  :datetime         not null
#  updated_at  :datetime         not null
#
# Indexes
#
#  index_number_formats_on_name  (name) UNIQUE
#

class NumberFormat < ApplicationRecord
end
