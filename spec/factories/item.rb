FactoryGirl.define do
  factory :item do
    name  'プリキュアのDVD'
    url   'http://example.com'
    image File.open("#{Rails.root}/db/data/image/work_1.png")
    main  true
  end
end
