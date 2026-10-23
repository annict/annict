# typed: false
# frozen_string_literal: true

RSpec.describe "GET /db/activities", type: :request do
  it "ユーザーがログインしていないとき、アクティビティリストが表示されること" do
    db_activity = create(:works_create_activity)
    work = db_activity.trackable

    get "/db/activities"

    expect(response.status).to eq(200)
    expect(response.body).to include(work.title)
  end

  it "ユーザーがログインしているとき、アクティビティリストが表示されること" do
    user = create(:registered_user)
    db_activity = create(:works_create_activity)
    work = db_activity.trackable

    login_as(user, scope: :user)

    get "/db/activities"

    expect(response.status).to eq(200)
    expect(response.body).to include(work.title)
  end

  it "ページネーションが正常に動作すること" do
    create(:works_create_activity)

    get "/db/activities", params: {page: 1}

    expect(response.status).to eq(200)
  end

  it "アクティビティが存在しないとき、空のページが表示されること" do
    get "/db/activities"

    expect(response.status).to eq(200)
  end

  it "無効なページパラメータが指定されたとき、正常に処理されること" do
    create(:works_create_activity)

    get "/db/activities", params: {page: "invalid"}

    expect(response.status).to eq(200)
  end

  it "非公開・削除済みの人物の編集履歴は表示されないこと" do
    published_person = create(:person, name: "公開中の人物")
    unpublished_person = create(:person, :unpublished, name: "非公開の人物")
    deleted_person = create(:person, :deleted, name: "削除済みの人物")
    [published_person, unpublished_person, deleted_person].each do |person|
      create(:db_activity, trackable: person, root_resource: person, action: "people.create", parameters: {new: person.attributes})
    end

    get "/db/activities"

    expect(response.status).to eq(200)
    expect(response.body).to include("公開中の人物")
    expect(response.body).not_to include("非公開の人物")
    expect(response.body).not_to include("削除済みの人物")
  end

  it "人物が非公開・削除済みのキャストの編集履歴は表示されないこと" do
    published_cast = create(:cast, person: create(:person, name: "公開中の人物"))
    unpublished_cast = create(:cast, person: create(:person, :unpublished, name: "非公開の人物"))
    deleted_cast = create(:cast, person: create(:person, :deleted, name: "削除済みの人物"))
    [published_cast, unpublished_cast, deleted_cast].each do |cast|
      create(:db_activity, trackable: cast, root_resource: cast.work, action: "casts.create", parameters: {new: cast.attributes})
    end

    get "/db/activities"

    expect(response.status).to eq(200)
    expect(response.body).to include("公開中の人物")
    expect(response.body).not_to include("非公開の人物")
    expect(response.body).not_to include("削除済みの人物")
  end

  it "人物が非公開・削除済みのスタッフの編集履歴は表示されず、団体のスタッフの編集履歴は表示されること" do
    published_staff = create(:staff, resource: create(:person, name: "公開中の人物"))
    unpublished_staff = create(:staff, resource: create(:person, :unpublished, name: "非公開の人物"))
    deleted_staff = create(:staff, resource: create(:person, :deleted, name: "削除済みの人物"))
    organization_staff = create(:staff, resource: create(:organization, name: "テスト団体"))
    [published_staff, unpublished_staff, deleted_staff, organization_staff].each do |staff|
      create(:db_activity, trackable: staff, root_resource: staff.work, action: "staffs.create", parameters: {new: staff.attributes})
    end

    get "/db/activities"

    expect(response.status).to eq(200)
    expect(response.body).to include("公開中の人物")
    expect(response.body).to include("テスト団体")
    expect(response.body).not_to include("非公開の人物")
    expect(response.body).not_to include("削除済みの人物")
  end
end
