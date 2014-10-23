class Tip < ActiveRecord::Base
  extend Enumerize

  enumerize :target, in: { new_user: 0, user: 1 }, scope: true
  enumerize :icon_name, in: { lightbulb_o: 0, bullhorn: 1 }
end
