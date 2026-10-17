# typed: false
# frozen_string_literal: true

RSpec.describe EmailNotificationMailer, type: :mailer do
  let(:user) { create(:registered_user) }
  let(:work) { create(:work) }

  before do
    published_cast_person = create(:person, name: "公開中の声優")
    unpublished_cast_person = create(:person, :unpublished, name: "非公開の声優")
    published_staff_person = create(:person, name: "公開中のスタッフ")
    deleted_staff_person = create(:person, :deleted, name: "削除済みのスタッフ")
    organization = create(:organization, name: "公開中の団体")

    create(:cast, work:, person: published_cast_person)
    create(:cast, work:, person: unpublished_cast_person)
    create(:staff, work:, resource: published_staff_person, name: published_staff_person.name)
    create(:staff, work:, resource: deleted_staff_person, name: deleted_staff_person.name)
    create(:staff, work:, resource: organization, name: organization.name)
    create(:person_favorite, user:, person: published_cast_person)
    create(:person_favorite, user:, person: unpublished_cast_person)
    create(:person_favorite, user:, person: deleted_staff_person)
  end

  it "お気に入り作品の追加通知から非公開・削除済みの人物名を除くこと" do
    body = described_class.favorite_works_added(user.id, [work.id]).body.decoded

    expect(body).to include("公開中の声優", "公開中のスタッフ", "公開中の団体")
    expect(body).not_to include("非公開の声優", "削除済みのスタッフ")

    # キャスト欄にも同じ名前が出るため、お気に入り欄の見出しより後ろだけで公開中のお気に入り人物を確認する
    favorites_section = body.split(I18n.t("email_notification_mailer.favorite_works_added.favorites_joins_work", locale: user.locale)).last
    expect(favorites_section).to include("公開中の声優")
  end

  it "関連作品の追加通知から非公開・削除済みの人物名を除くこと" do
    body = described_class.related_works_added(user.id, [work.id]).body.decoded

    expect(body).to include("公開中の声優", "公開中のスタッフ", "公開中の団体")
    expect(body).not_to include("非公開の声優", "削除済みのスタッフ")
  end
end
