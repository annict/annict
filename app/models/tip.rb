# frozen_string_literal: true
# == Schema Information
#
# Table name: tips
#
#  id         :integer          not null, primary key
#  target     :integer          not null
#  slug       :string           not null
#  title      :string           not null
#  icon_name  :string           not null
#  created_at :datetime         not null
#  updated_at :datetime         not null
#  title_en   :string           default(""), not null
#
# Indexes
#
#  index_tips_on_slug  (slug) UNIQUE
#

class Tip < ActiveRecord::Base
  extend Enumerize

  enumerize :target, in: { new_user: 0, user: 1 }, scope: true
end
