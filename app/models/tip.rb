# frozen_string_literal: true
# == Schema Information
#
# Table name: tips
#
#  id         :bigint           not null, primary key
#  icon_name  :string(255)      not null
#  locale     :string           default("other"), not null
#  slug       :string(255)      not null
#  target     :integer          not null
#  title      :string(255)      not null
#  created_at :datetime         not null
#  updated_at :datetime         not null
#
# Indexes
#
#  index_tips_on_locale           (locale)
#  index_tips_on_slug_and_locale  (slug,locale) UNIQUE
#

class Tip < ApplicationRecord
  extend Enumerize

  include Localizable

  enumerize :target, in: { new_user: 0, user: 1 }, scope: true
end
