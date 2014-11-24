FactoryGirl.define do
  factory :cover_image do
    association :work
    file_name 'cover-image-1.jpg'
    location '○○県△△市 □□□'
  end
end
