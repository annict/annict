# typed: false
# frozen_string_literal: true

RSpec.describe "GET /works/newest", type: :request do
  it "アクセスできること" do
    work = create(:work)

    get "/works/newest"

    expect(response.status).to eq(200)
    expect(response.body).to include(work.title)
  end

  it "削除されていない作品のみを表示すること" do
    kept_work = create(:work)
    deleted_work = create(:work, deleted_at: Time.current)

    get "/works/newest"

    expect(response.body).to include(kept_work.title)
    expect(response.body).not_to include(deleted_work.title)
  end

  it "新しい作品から順に表示すること" do
    old_work = create(:work)
    new_work = create(:work)

    get "/works/newest"

    expect(response.body.index(new_work.title)).to be < response.body.index(old_work.title)
  end

  it "グリッド表示（デフォルト）で30件表示すること" do
    create_list(:work, 31)

    get "/works/newest"

    # 各作品のリンクでカウント
    expect(response.body.scan(%r{<div class="c-work-card}).count).to eq(30)
  end

  it "display=gridで30件表示すること" do
    create_list(:work, 31)

    get "/works/newest", params: {display: "grid"}

    expect(response.body.scan(%r{<div class="c-work-card}).count).to eq(30)
  end

  it "display=grid_smallで120件表示すること" do
    create_list(:work, 121)

    get "/works/newest", params: {display: "grid_small"}

    expect(response.body.scan(%r{<div class="c-work-grid__col}).count).to eq(120)
  end

  it "無効なdisplayパラメータの場合はグリッド表示（30件）にすること" do
    create_list(:work, 31)

    get "/works/newest", params: {display: "invalid"}

    expect(response.body.scan(%r{<div class="c-work-card}).count).to eq(30)
  end

  it "ページネーションが動作すること" do
    create_list(:work, 35)

    get "/works/newest", params: {page: 2}

    expect(response.status).to eq(200)
    expect(response.body.scan(%r{<div class="c-work-card}).count).to eq(5)
  end

  it "グリッド表示のときはキャストとスタッフ情報を含むこと" do
    work = create(:work)
    character = create(:character)
    person = create(:person)
    cast = create(:cast, work: work, character: character, person: person)
    staff = create(:staff, work: work, resource: person, role: "original_creator")

    # キャストとスタッフが正しく作成され、publishedであることを確認
    expect(cast.unpublished_at).to be_nil
    expect(staff.unpublished_at).to be_nil
    expect(Cast.only_kept.where(work: work)).to include(cast)
    expect(Staff.only_kept.where(work: work)).to include(staff)

    get "/works/newest"

    # 作品が表示されていることを確認
    expect(response.body).to include(work.title)

    # キャスト・スタッフ情報が表示されていることを確認
    expect(response.body).to include(character.name)
    expect(response.body).to include(person.name)
  end

  it "grid_small表示のときはキャストとスタッフ情報を含まないこと" do
    work = create(:work)
    character = create(:character)
    person = create(:person)
    cast = create(:cast, work: work, character: character, person: person)
    staff = create(:staff, work: work, resource: person, role: "original_creator")

    get "/works/newest", params: {display: "grid_small"}

    expect(response.body).to include(work.title)
    expect(response.body).not_to include(character.name)
    expect(response.body).not_to include(person.name)
  end
end
