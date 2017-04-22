# frozen_string_literal: true

namespace :tmp do
  task set_userland_categories: :environment do
    categories = [
      { name: "Webアプリ", name_en: "Web App", sort_number: 100 },
      { name: "iOSアプリ", name_en: "iOS App", sort_number: 200 },
      { name: "Androidアプリ", name_en: "Android App", sort_number: 300 },
      { name: "ツール", name_en: "Tool", sort_number: 400 },
      { name: "開発者向けライブラリ", name_en: "Library for Developers", sort_number: 500 },
      { name: "その他", name_en: "Other", sort_number: 1000 }
    ]

    UserlandCategory.create(categories)
  end
end
