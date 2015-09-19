namespace :tmp do
  task delete_sub_items: :environment do
    Item.where(main: false).delete_all
  end
end
