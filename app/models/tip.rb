# frozen_string_literal: true
# == Schema Information
#
# Table name: tips
#
#  id         :integer          not null, primary key
#  target     :integer          not null
#  slug       :string(255)      not null
#  title      :string(255)      not null
#  icon_name  :string(255)      not null
#  created_at :datetime         not null
#  updated_at :datetime         not null
#  locale     :string(255)      default("other"), not null
#
# Indexes
#
#  index_tips_on_slug_and_locale  (slug,locale) UNIQUE
#

class Tip < ApplicationRecord
  extend Enumerize

  enumerize :target, in: { new_user: 0, user: 1 }, scope: true
end
