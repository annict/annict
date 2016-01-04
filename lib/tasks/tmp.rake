namespace :tmp do
  task create_data: :environment do
    PREFECTURES.each do |prefecture_name|
      prefecture = Prefecture.where(name: prefecture_name).first_or_create!
      puts "prefecture: #{prefecture.name}"
    end

    works = Work.where.not(sc_tid: nil).order(:id)
    works = works.where("id >= ?", 1)
    works.find_each do |work|
      puts "work_id: #{work.id}"
      doc = Nokogiri::XML(open("http://cal.syoboi.jp/db.php?TID=#{work.sc_tid}&Command=TitleLookup"))
      doc.css("Comment").each do |item|
        begin
          comment = item.text + "*"
          comment = comment.
            gsub(/\n/, ",").
            gsub(":,", ",").
            gsub(";", ":").
            gsub("：", ":")

          save_staff_info(work, comment)
          save_cast_info(work, comment)
        rescue => e
          binding.pry
        end
      end
    end
  end
end

def save_staff_info(work, comment)
  staff_str = comment[/\*(スタッフ.+?)\*/, 1]

  if staff_str.present?
    staff_ary = staff_str.split(",").select { |str| str.include?(":") }
    staffs = staff_ary.map { |str| str.split(/:(.*):/).select(&:present?) }
    staffs = staffs[0..-2] if work.id == 2121
    staffs = staffs.to_h
    org_str = staffs["アニメーション制作"] ||
              staffs["制作"] ||
              staffs["制作スタジオ"] ||
              staffs["アニメーション製作"] ||
              staffs["アニメ制作"]
    orgs = org_str.try!(:split, "、")
    if orgs.present?
      orgs.each do |org|
        organization = Organization.where(name: org).first_or_create!
        work.work_organizations.where(organization: organization, role: :producer, sort_number: 10).first_or_create!
      end
    end

    staffs = staffs.except("アニメーション制作", "制作", "制作スタジオ", "アニメーション製作", "アニメ制作")

    work_staff_roles = Staff.role.values.map do |role_value|
      { value: role_value, text: I18n.t("enumerize.staff.role.#{role_value}") }
    end
    i = 1
    work_staff_roles.each do |wsp_role|
      name = staffs[wsp_role[:text]]
      next if name.blank?
      person = Person.where(name: name).first_or_create!
      work.staffs.where(person: person, name: person.name, role: wsp_role[:value]).first_or_create! do |staff|
        staff.sort_number = i * 10
        i += 1
      end
    end

    staffs = staffs.except(*(work_staff_roles.map { |wspr| wspr[:text] }))

    staffs.each do |role, name|
      person = Person.where(name: name).first_or_create!
      work.staffs.where(person: person, name: person.name, role: :other, role_other: role).first_or_create! do |staff|
        staff.sort_number = 100
      end
    end
  end
end

def save_cast_info(work, comment)
  cast_str = comment[/\*(キャスト.+?)\*/, 1]

  if cast_str.present?
    cast_ary = cast_str.split(",").select { |str| str.include?(":") }
    casts = cast_ary.map { |str| str.split(/:(.*):/).select(&:present?) }
    casts = casts.select { |cast| cast.length == 2 }.to_h

    i = 1
    casts.each do |part, name|
      person = Person.where(name: name).first_or_create!
      work.casts.where(person: person, name: person.name, part: part).first_or_create! do |cast|
        cast.sort_number = i * 10
        i += 1
      end
    end
  end
end

PREFECTURES = %w(
  北海道
  青森県
  岩手県
  宮城県
  秋田県
  山形県
  福島県
  茨城県
  栃木県
  群馬県
  埼玉県
  千葉県
  東京都
  神奈川県
  新潟県
  富山県
  石川県
  福井県
  山梨県
  長野県
  岐阜県
  静岡県
  愛知県
  三重県
  滋賀県
  京都府
  大阪府
  兵庫県
  奈良県
  和歌山県
  鳥取県
  島根県
  岡山県
  広島県
  山口県
  徳島県
  香川県
  愛媛県
  高知県
  福岡県
  佐賀県
  長崎県
  熊本県
  大分県
  宮崎県
  鹿児島県
  沖縄県
  海外
)
