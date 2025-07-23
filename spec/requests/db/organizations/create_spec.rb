# typed: false
# frozen_string_literal: true

RSpec.describe "POST /db/organizations", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    organization_params = {
      rows: "御三家,ごさんけ"
    }

    post "/db/organizations", params: {deprecated_db_organization_rows_form: organization_params}

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(Organization.all.size).to eq(0)
  end

  it "編集者権限のないユーザーがログインしているとき、アクセスできないこと" do
    user = create(:registered_user)
    organization_params = {
      rows: "御三家,ごさんけ"
    }

    login_as(user, scope: :user)

    post "/db/organizations", params: {deprecated_db_organization_rows_form: organization_params}

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(Organization.all.size).to eq(0)
  end

  it "編集者権限のあるユーザーがログインしているとき、組織を登録できること" do
    user = create(:registered_user, :with_editor_role)
    organization_params = {
      rows: "御三家,ごさんけ"
    }

    login_as(user, scope: :user)

    expect(Organization.all.size).to eq(0)

    post "/db/organizations", params: {deprecated_db_organization_rows_form: organization_params}

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("登録しました")
    expect(Organization.all.size).to eq(1)

    organization = Organization.first
    expect(organization.name).to eq("御三家")
    expect(organization.name_kana).to eq("ごさんけ")
  end

  it "編集者権限のあるユーザーがログインしているとき、複数行の組織データを一度に登録できること" do
    user = create(:registered_user, :with_editor_role)
    organization_params = {
      rows: "サンライズ,さんらいず\n東映アニメーション,とうえいあにめーしょん"
    }

    login_as(user, scope: :user)

    post "/db/organizations", params: {deprecated_db_organization_rows_form: organization_params}

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("登録しました")
    expect(Organization.all.size).to eq(2)

    organizations = Organization.order(:created_at)
    expect(organizations[0].name).to eq("サンライズ")
    expect(organizations[0].name_kana).to eq("さんらいず")
    expect(organizations[1].name).to eq("東映アニメーション")
    expect(organizations[1].name_kana).to eq("とうえいあにめーしょん")
  end

  it "編集者権限のあるユーザーがログインしているとき、rowsが空の場合バリデーションエラーになること" do
    user = create(:registered_user, :with_editor_role)
    organization_params = {
      rows: ""
    }

    login_as(user, scope: :user)

    post "/db/organizations", params: {deprecated_db_organization_rows_form: organization_params}

    expect(response.status).to eq(422)
    expect(Organization.all.size).to eq(0)
  end

  it "編集者権限のあるユーザーがログインしているとき、nameかなを省略できること" do
    user = create(:registered_user, :with_editor_role)
    organization_params = {
      rows: "A-1 Pictures"
    }

    login_as(user, scope: :user)

    post "/db/organizations", params: {deprecated_db_organization_rows_form: organization_params}

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("登録しました")
    expect(Organization.all.size).to eq(1)

    organization = Organization.first
    expect(organization.name).to eq("A-1 Pictures")
    expect(organization.name_kana).to eq("")
  end

  it "編集者権限のあるユーザーがログインしているとき、不正なフォーマットの場合バリデーションエラーになること" do
    user = create(:registered_user, :with_editor_role)
    organization_params = {
      rows: "組織名,かな,余分な列"
    }

    login_as(user, scope: :user)

    post "/db/organizations", params: {deprecated_db_organization_rows_form: organization_params}

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("登録しました")
    expect(Organization.all.size).to eq(1)

    # 余分な列は無視される
    organization = Organization.first
    expect(organization.name).to eq("組織名")
    expect(organization.name_kana).to eq("かな")
  end
end
