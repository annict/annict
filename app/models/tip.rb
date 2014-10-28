# == Schema Information
#
# Table name: tips
#
#  id           :integer          not null, primary key
#  target       :integer          not null
#  partial_name :string(255)      not null
#  title        :string(255)      not null
#  icon_name    :string(255)      not null
#  created_at   :datetime         not null
#  updated_at   :datetime         not null
#
# Indexes
#
#  index_tips_on_partial_name  (partial_name) UNIQUE
#

class Tip < ActiveRecord::Base
  extend Enumerize

  enumerize :target, in: { new_user: 0, user: 1 }, scope: true
  enumerize :icon_name, in: { lightbulb_o: 0, bullhorn: 1 }
end
