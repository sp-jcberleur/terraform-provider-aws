// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestIPProtocol(t *testing.T) {
	ctx := acctest.Context(t)
	t.Parallel()

	type testCase struct {
		val1, val2 tfec2.IPProtocol
		equals     bool
	}
	tests := map[string]testCase{
		"name, number (equivalent)": {
			val1:   tfec2.IPProtocol{StringValue: types.StringValue("icmp")},
			val2:   tfec2.IPProtocol{StringValue: types.StringValue(acctest.Ct1)},
			equals: true,
		},
		"number, name (equivalent)": {
			val1:   tfec2.IPProtocol{StringValue: types.StringValue(acctest.Ct1)},
			val2:   tfec2.IPProtocol{StringValue: types.StringValue("icmp")},
			equals: true,
		},
		"name, number (not equivalent)": {
			val1:   tfec2.IPProtocol{StringValue: types.StringValue("icmp")},
			val2:   tfec2.IPProtocol{StringValue: types.StringValue(acctest.Ct2)},
			equals: false,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			equals, _ := test.val1.StringSemanticEquals(ctx, test.val2)

			if got, want := equals, test.equals; got != want {
				t.Errorf("StringSemanticEquals(%q, %q) = %v, want %v", test.val1, test.val2, got, want)
			}
		})
	}
}

func TestAccVPCSecurityGroupIngressRule_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v ec2.SecurityGroupRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_security_group_ingress_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupIngressRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupIngressRuleConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "cidr_ipv4", "10.0.0.0/8"),
					resource.TestCheckNoResourceAttr(resourceName, "cidr_ipv6"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, "from_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "ip_protocol", "tcp"),
					resource.TestCheckNoResourceAttr(resourceName, "prefix_list_id"),
					resource.TestCheckNoResourceAttr(resourceName, "referenced_security_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_rule_id"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrTags),
					resource.TestCheckResourceAttr(resourceName, "to_port", "8080"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSecurityGroupIngressRule_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v ec2.SecurityGroupRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_security_group_ingress_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupIngressRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupIngressRuleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(ctx, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfec2.ResourceSecurityGroupIngressRule, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCSecurityGroupIngressRule_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v ec2.SecurityGroupRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_security_group_ingress_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupIngressRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupIngressRuleConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCSecurityGroupIngressRuleConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccVPCSecurityGroupIngressRuleConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccVPCSecurityGroupIngressRule_tags_computed(t *testing.T) {
	ctx := acctest.Context(t)
	var v ec2.SecurityGroupRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_security_group_ingress_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupIngressRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupIngressRuleConfig_computed(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "tags.eip"),
				),
			},
		},
	})
}

func TestAccVPCSecurityGroupIngressRule_DefaultTags_providerOnly(t *testing.T) {
	ctx := acctest.Context(t)
	var v ec2.SecurityGroupRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_security_group_ingress_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupIngressRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1(acctest.CtProviderKey1, acctest.CtProviderValue1),
					testAccVPCSecurityGroupIngressRuleConfig_basic(rName),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey1", acctest.CtProviderValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags2(acctest.CtProviderKey1, acctest.CtProviderValue1, "providerkey2", "providervalue2"),
					testAccVPCSecurityGroupIngressRuleConfig_basic(rName),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey1", acctest.CtProviderValue1),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey2", "providervalue2"),
				),
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1(acctest.CtProviderKey1, acctest.CtValue1),
					testAccVPCSecurityGroupIngressRuleConfig_basic(rName),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey1", acctest.CtValue1),
				),
			},
		},
	})
}

func TestAccVPCSecurityGroupIngressRule_DefaultTags_updateToProviderOnly(t *testing.T) {
	ctx := acctest.Context(t)
	var v ec2.SecurityGroupRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_security_group_ingress_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupIngressRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupIngressRuleConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", acctest.CtValue1),
				),
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1(acctest.CtKey1, acctest.CtValue1),
					testAccVPCSecurityGroupIngressRuleConfig_basic(rName),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSecurityGroupIngressRule_DefaultTags_updateToResourceOnly(t *testing.T) {
	ctx := acctest.Context(t)
	var v ec2.SecurityGroupRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_security_group_ingress_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupIngressRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1(acctest.CtKey1, acctest.CtValue1),
					testAccVPCSecurityGroupIngressRuleConfig_basic(rName),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", acctest.CtValue1),
				),
			},
			{
				Config: testAccVPCSecurityGroupIngressRuleConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSecurityGroupIngressRule_DefaultTagsProviderAndResource_nonOverlappingTag(t *testing.T) {
	ctx := acctest.Context(t)
	var v ec2.SecurityGroupRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_security_group_ingress_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupIngressRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1(acctest.CtProviderKey1, acctest.CtProviderValue1),
					testAccVPCSecurityGroupIngressRuleConfig_tags1(rName, acctest.CtResourceKey1, acctest.CtResourceValue1),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags.resourcekey1", acctest.CtResourceValue1),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey1", acctest.CtProviderValue1),
					resource.TestCheckResourceAttr(resourceName, "tags_all.resourcekey1", acctest.CtResourceValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1(acctest.CtProviderKey1, acctest.CtProviderValue1),
					testAccVPCSecurityGroupIngressRuleConfig_tags2(rName, acctest.CtResourceKey1, acctest.CtResourceValue1, "resourcekey2", acctest.CtResourceValue2),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "tags.resourcekey1", acctest.CtResourceValue1),
					resource.TestCheckResourceAttr(resourceName, "tags.resourcekey2", acctest.CtResourceValue2),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey1", acctest.CtProviderValue1),
					resource.TestCheckResourceAttr(resourceName, "tags_all.resourcekey1", acctest.CtResourceValue1),
					resource.TestCheckResourceAttr(resourceName, "tags_all.resourcekey2", acctest.CtResourceValue2),
				),
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1("providerkey2", "providervalue2"),
					testAccVPCSecurityGroupIngressRuleConfig_tags1(rName, "resourcekey3", "resourcevalue3"),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags.resourcekey3", "resourcevalue3"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey2", "providervalue2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.resourcekey3", "resourcevalue3"),
				),
			},
		},
	})
}

func TestAccVPCSecurityGroupIngressRule_DefaultTagsProviderAndResource_overlappingTag(t *testing.T) {
	ctx := acctest.Context(t)
	var v ec2.SecurityGroupRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_security_group_ingress_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupIngressRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1(acctest.CtOverlapKey1, acctest.CtProviderValue1),
					testAccVPCSecurityGroupIngressRuleConfig_tags1(rName, acctest.CtOverlapKey1, acctest.CtResourceValue1),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.overlapkey1", acctest.CtResourceValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags2(acctest.CtOverlapKey1, acctest.CtProviderValue1, "overlapkey2", "providervalue2"),
					testAccVPCSecurityGroupIngressRuleConfig_tags2(rName, acctest.CtOverlapKey1, acctest.CtResourceValue1, "overlapkey2", acctest.CtResourceValue2),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags.overlapkey1", acctest.CtResourceValue1),
					resource.TestCheckResourceAttr(resourceName, "tags.overlapkey2", acctest.CtResourceValue2),
					resource.TestCheckResourceAttr(resourceName, "tags_all.overlapkey1", acctest.CtResourceValue1),
					resource.TestCheckResourceAttr(resourceName, "tags_all.overlapkey2", acctest.CtResourceValue2),
				),
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1(acctest.CtOverlapKey1, acctest.CtProviderValue1),
					testAccVPCSecurityGroupIngressRuleConfig_tags1(rName, acctest.CtOverlapKey1, acctest.CtResourceValue2),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.overlapkey1", acctest.CtResourceValue2),
					resource.TestCheckResourceAttr(resourceName, "tags_all.overlapkey1", acctest.CtResourceValue2),
				),
			},
		},
	})
}

func TestAccVPCSecurityGroupIngressRule_DefaultTagsProviderAndResource_duplicateTag(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var v ec2.SecurityGroupRule
	resourceName := "aws_vpc_security_group_ingress_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupIngressRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1("overlapkey", "overlapvalue"),
					testAccVPCSecurityGroupIngressRuleConfig_tags1(rName, "overlapkey", "overlapvalue"),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
				),
			},
		},
	})
}

func TestAccVPCSecurityGroupIngressRule_updateTagsKnownAtApply(t *testing.T) {
	ctx := acctest.Context(t)
	var v ec2.SecurityGroupRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_security_group_ingress_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupIngressRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupIngressRuleConfig_tagsComputedFromDataSource1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
				),
			},
			{
				Config: testAccVPCSecurityGroupIngressRuleConfig_tagsComputedFromDataSource2(rName, acctest.CtKey1, acctest.CtValue1, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSecurityGroupIngressRule_defaultAndIgnoreTags(t *testing.T) {
	ctx := acctest.Context(t)
	var v ec2.SecurityGroupRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_security_group_ingress_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupIngressRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupIngressRuleConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(ctx, resourceName, &v),
					testAccCheckSecurityGroupIngressRuleUpdateTags(ctx, &v, nil, map[string]string{"defaultkey1": "defaultvalue1"}),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultAndIgnoreTagsKeyPrefixes1("defaultkey1", "defaultvalue1", "defaultkey"),
					testAccVPCSecurityGroupIngressRuleConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				),
				PlanOnly: true,
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultAndIgnoreTagsKeys1("defaultkey1", "defaultvalue1"),
					testAccVPCSecurityGroupIngressRuleConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				),
				PlanOnly: true,
			},
		},
	})
}

func TestAccVPCSecurityGroupIngressRule_ignoreTags(t *testing.T) {
	ctx := acctest.Context(t)
	var v ec2.SecurityGroupRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_security_group_ingress_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupIngressRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupIngressRuleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(ctx, resourceName, &v),
					testAccCheckSecurityGroupIngressRuleUpdateTags(ctx, &v, nil, map[string]string{"ignorekey1": "ignorevalue1"}),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigIgnoreTagsKeyPrefixes1("ignorekey"),
					testAccVPCSecurityGroupIngressRuleConfig_basic(rName),
				),
				PlanOnly: true,
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigIgnoreTagsKeys("ignorekey1"),
					testAccVPCSecurityGroupIngressRuleConfig_basic(rName),
				),
				PlanOnly: true,
			},
		},
	})
}

func TestAccVPCSecurityGroupIngressRule_cidrIPv4(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 ec2.SecurityGroupRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_security_group_ingress_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupIngressRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupIngressRuleConfig_cidrIPv4(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "cidr_ipv4", "0.0.0.0/0"),
					resource.TestCheckNoResourceAttr(resourceName, "cidr_ipv6"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, "from_port", "53"),
					resource.TestCheckResourceAttr(resourceName, "ip_protocol", "udp"),
					resource.TestCheckNoResourceAttr(resourceName, "prefix_list_id"),
					resource.TestCheckNoResourceAttr(resourceName, "referenced_security_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_rule_id"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrTags),
					resource.TestCheckResourceAttr(resourceName, "to_port", "53"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCSecurityGroupIngressRuleConfig_cidrIPv4Updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(ctx, resourceName, &v2),
					testAccCheckSecurityGroupRuleNotRecreated(&v2, &v1),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "cidr_ipv4", "10.0.0.0/16"),
					resource.TestCheckNoResourceAttr(resourceName, "cidr_ipv6"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, "from_port", "-1"),
					resource.TestCheckResourceAttr(resourceName, "ip_protocol", acctest.Ct1),
					resource.TestCheckNoResourceAttr(resourceName, "prefix_list_id"),
					resource.TestCheckNoResourceAttr(resourceName, "referenced_security_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_rule_id"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrTags),
					resource.TestCheckResourceAttr(resourceName, "to_port", "-1"),
				),
			},
		},
	})
}

func TestAccVPCSecurityGroupIngressRule_cidrIPv6(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 ec2.SecurityGroupRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_security_group_ingress_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupIngressRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupIngressRuleConfig_cidrIPv6(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "cidr_ipv4"),
					resource.TestCheckResourceAttr(resourceName, "cidr_ipv6", "2001:db8:85a3::/64"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, "from_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "ip_protocol", "tcp"),
					resource.TestCheckNoResourceAttr(resourceName, "prefix_list_id"),
					resource.TestCheckNoResourceAttr(resourceName, "referenced_security_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_rule_id"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrTags),
					resource.TestCheckResourceAttr(resourceName, "to_port", "8080"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCSecurityGroupIngressRuleConfig_cidrIPv6Updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(ctx, resourceName, &v2),
					testAccCheckSecurityGroupRuleNotRecreated(&v2, &v1),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "cidr_ipv4"),
					resource.TestCheckResourceAttr(resourceName, "cidr_ipv6", "2001:db8:85a3:2::/64"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
					resource.TestCheckNoResourceAttr(resourceName, "from_port"),
					resource.TestCheckResourceAttr(resourceName, "ip_protocol", "icmpv6"),
					resource.TestCheckNoResourceAttr(resourceName, "prefix_list_id"),
					resource.TestCheckNoResourceAttr(resourceName, "referenced_security_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_rule_id"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrTags),
					resource.TestCheckNoResourceAttr(resourceName, "to_port"),
				),
			},
		},
	})
}

func TestAccVPCSecurityGroupIngressRule_description(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 ec2.SecurityGroupRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_security_group_ingress_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupIngressRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupIngressRuleConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCSecurityGroupIngressRuleConfig_description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(ctx, resourceName, &v2),
					testAccCheckSecurityGroupRuleNotRecreated(&v2, &v1),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description2"),
				),
			},
		},
	})
}

func TestAccVPCSecurityGroupIngressRule_prefixListID(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 ec2.SecurityGroupRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_security_group_ingress_rule.test"
	vpcEndpoint1ResourceName := "aws_vpc_endpoint.test1"
	vpcEndpoint2ResourceName := "aws_vpc_endpoint.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupIngressRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupIngressRuleConfig_prefixListID(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "cidr_ipv4"),
					resource.TestCheckNoResourceAttr(resourceName, "cidr_ipv6"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, "from_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "ip_protocol", "tcp"),
					resource.TestCheckResourceAttrPair(resourceName, "prefix_list_id", vpcEndpoint1ResourceName, "prefix_list_id"),
					resource.TestCheckNoResourceAttr(resourceName, "referenced_security_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_rule_id"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrTags),
					resource.TestCheckResourceAttr(resourceName, "to_port", "8080"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCSecurityGroupIngressRuleConfig_prefixListIDUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(ctx, resourceName, &v2),
					testAccCheckSecurityGroupRuleNotRecreated(&v2, &v1),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "cidr_ipv4"),
					resource.TestCheckNoResourceAttr(resourceName, "cidr_ipv6"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, "from_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "ip_protocol", "tcp"),
					resource.TestCheckResourceAttrPair(resourceName, "prefix_list_id", vpcEndpoint2ResourceName, "prefix_list_id"),
					resource.TestCheckNoResourceAttr(resourceName, "referenced_security_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_rule_id"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrTags),
					resource.TestCheckResourceAttr(resourceName, "to_port", "8080"),
				),
			},
		},
	})
}

func TestAccVPCSecurityGroupIngressRule_referencedSecurityGroupID(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 ec2.SecurityGroupRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_security_group_ingress_rule.test"
	securityGroup1ResourceName := "aws_security_group.test"
	securityGroup2ResourceName := "aws_security_group.test1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupIngressRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupIngressRuleConfig_referencedSecurityGroupID(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "cidr_ipv4"),
					resource.TestCheckNoResourceAttr(resourceName, "cidr_ipv6"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, "from_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "ip_protocol", "tcp"),
					resource.TestCheckNoResourceAttr(resourceName, "prefix_list_id"),
					resource.TestCheckResourceAttrPair(resourceName, "referenced_security_group_id", securityGroup1ResourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_rule_id"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrTags),
					resource.TestCheckResourceAttr(resourceName, "to_port", "8080"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCSecurityGroupIngressRuleConfig_referencedSecurityGroupIDUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(ctx, resourceName, &v2),
					testAccCheckSecurityGroupRuleNotRecreated(&v2, &v1),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "cidr_ipv4"),
					resource.TestCheckNoResourceAttr(resourceName, "cidr_ipv6"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, "from_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "ip_protocol", "tcp"),
					resource.TestCheckNoResourceAttr(resourceName, "prefix_list_id"),
					resource.TestCheckResourceAttrPair(resourceName, "referenced_security_group_id", securityGroup2ResourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_rule_id"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrTags),
					resource.TestCheckResourceAttr(resourceName, "to_port", "8080"),
				),
			},
		},
	})
}

func TestAccVPCSecurityGroupIngressRule_ReferencedSecurityGroupID_peerVPC(t *testing.T) {
	ctx := acctest.Context(t)
	var v ec2.SecurityGroupRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_security_group_ingress_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckSecurityGroupIngressRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupIngressRuleConfig_referencedSecurityGroupIDPeerVPC(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "cidr_ipv4"),
					resource.TestCheckNoResourceAttr(resourceName, "cidr_ipv6"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, "from_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "ip_protocol", "tcp"),
					resource.TestCheckNoResourceAttr(resourceName, "prefix_list_id"),
					resource.TestMatchResourceAttr(resourceName, "referenced_security_group_id", regexache.MustCompile("^[0-9]{12}/sg-[0-9a-z]{17}$")),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_rule_id"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrTags),
					resource.TestCheckResourceAttr(resourceName, "to_port", "8080"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSecurityGroupIngressRule_updateSourceType(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 ec2.SecurityGroupRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_security_group_ingress_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupIngressRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupIngressRuleConfig_cidrIPv4(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "cidr_ipv4", "0.0.0.0/0"),
					resource.TestCheckNoResourceAttr(resourceName, "cidr_ipv6"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, "from_port", "53"),
					resource.TestCheckResourceAttr(resourceName, "ip_protocol", "udp"),
					resource.TestCheckNoResourceAttr(resourceName, "prefix_list_id"),
					resource.TestCheckNoResourceAttr(resourceName, "referenced_security_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_rule_id"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrTags),
					resource.TestCheckResourceAttr(resourceName, "to_port", "53"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCSecurityGroupIngressRuleConfig_cidrIPv6(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(ctx, resourceName, &v2),
					testAccCheckSecurityGroupRuleRecreated(&v2, &v1),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "cidr_ipv4"),
					resource.TestCheckResourceAttr(resourceName, "cidr_ipv6", "2001:db8:85a3::/64"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, "from_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "ip_protocol", "tcp"),
					resource.TestCheckNoResourceAttr(resourceName, "prefix_list_id"),
					resource.TestCheckNoResourceAttr(resourceName, "referenced_security_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_rule_id"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrTags),
					resource.TestCheckResourceAttr(resourceName, "to_port", "8080"),
				),
			},
		},
	})
}

func testAccCheckSecurityGroupRuleNotRecreated(i, j *ec2.SecurityGroupRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.SecurityGroupRuleId) != aws.StringValue(j.SecurityGroupRuleId) {
			return errors.New("VPC Security Group Rule was recreated")
		}

		return nil
	}
}

func testAccCheckSecurityGroupRuleRecreated(i, j *ec2.SecurityGroupRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.SecurityGroupRuleId) == aws.StringValue(j.SecurityGroupRuleId) {
			return errors.New("VPC Security Group Rule was not recreated")
		}

		return nil
	}
}

func testAccCheckSecurityGroupIngressRuleDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpc_security_group_ingress_rule" {
				continue
			}

			_, err := tfec2.FindSecurityGroupIngressRuleByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("VPC Security Group Ingress Rule still exists: %s", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckSecurityGroupIngressRuleExists(ctx context.Context, n string, v *ec2.SecurityGroupRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No VPC Security Group Ingress Rule ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn(ctx)

		output, err := tfec2.FindSecurityGroupIngressRuleByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckSecurityGroupIngressRuleUpdateTags(ctx context.Context, v *ec2.SecurityGroupRule, oldTags, newTags map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn(ctx)

		return tfec2.UpdateTags(ctx, conn, aws.StringValue(v.SecurityGroupRuleId), oldTags, newTags)
	}
}

func testAccVPCSecurityGroupRuleConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
  name   = %[1]q

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCSecurityGroupIngressRuleConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccVPCSecurityGroupRuleConfig_base(rName), `
resource "aws_vpc_security_group_ingress_rule" "test" {
  security_group_id = aws_security_group.test.id

  cidr_ipv4   = "10.0.0.0/8"
  from_port   = 80
  ip_protocol = "tcp"
  to_port     = 8080
}
`)
}

func testAccVPCSecurityGroupIngressRuleConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccVPCSecurityGroupRuleConfig_base(rName), fmt.Sprintf(`
resource "aws_vpc_security_group_ingress_rule" "test" {
  security_group_id = aws_security_group.test.id

  cidr_ipv4   = "10.0.0.0/8"
  from_port   = 80
  ip_protocol = "tcp"
  to_port     = 8080

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccVPCSecurityGroupIngressRuleConfig_computed(rName string) string {
	return acctest.ConfigCompose(testAccVPCSecurityGroupRuleConfig_base(rName), `
resource "aws_eip" "test" {
  domain = "vpc"
}

resource "aws_vpc_security_group_ingress_rule" "test" {
  security_group_id = aws_security_group.test.id

  cidr_ipv4   = "10.0.0.0/8"
  from_port   = 80
  ip_protocol = "tcp"
  to_port     = 8080

  tags = {
    eip = aws_eip.test.public_ip
  }
}
`)
}

func testAccVPCSecurityGroupIngressRuleConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccVPCSecurityGroupRuleConfig_base(rName), fmt.Sprintf(`
resource "aws_vpc_security_group_ingress_rule" "test" {
  security_group_id = aws_security_group.test.id

  cidr_ipv4   = "10.0.0.0/8"
  from_port   = 80
  ip_protocol = "tcp"
  to_port     = 8080

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccVPCSecurityGroupIngressRuleConfig_computedTagsBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
  tags       = local.tags
}

data "aws_vpc" "test" {
  id = aws_vpc.test.id
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
  name   = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_security_group_ingress_rule" "test" {
  security_group_id = aws_security_group.test.id

  cidr_ipv4   = "10.0.0.0/8"
  from_port   = 80
  ip_protocol = "tcp"
  to_port     = 8080

  tags = data.aws_vpc.test.tags
}
`, rName)
}

func testAccVPCSecurityGroupIngressRuleConfig_tagsComputedFromDataSource1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccVPCSecurityGroupIngressRuleConfig_computedTagsBase(rName), fmt.Sprintf(`
locals {
  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccVPCSecurityGroupIngressRuleConfig_tagsComputedFromDataSource2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccVPCSecurityGroupIngressRuleConfig_computedTagsBase(rName), fmt.Sprintf(`
locals {
  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccVPCSecurityGroupIngressRuleConfig_cidrIPv4(rName string) string {
	return acctest.ConfigCompose(testAccVPCSecurityGroupRuleConfig_base(rName), `
resource "aws_vpc_security_group_ingress_rule" "test" {
  security_group_id = aws_security_group.test.id

  cidr_ipv4   = "0.0.0.0/0"
  from_port   = 53
  ip_protocol = "udp"
  to_port     = 53
}
`)
}

func testAccVPCSecurityGroupIngressRuleConfig_cidrIPv4Updated(rName string) string {
	return acctest.ConfigCompose(testAccVPCSecurityGroupRuleConfig_base(rName), `
resource "aws_vpc_security_group_ingress_rule" "test" {
  security_group_id = aws_security_group.test.id

  cidr_ipv4   = "10.0.0.0/16"
  from_port   = -1
  ip_protocol = "1"
  to_port     = -1
}
`)
}

func testAccVPCSecurityGroupIngressRuleConfig_cidrIPv6(rName string) string {
	return acctest.ConfigCompose(testAccVPCSecurityGroupRuleConfig_base(rName), `
resource "aws_vpc_security_group_ingress_rule" "test" {
  security_group_id = aws_security_group.test.id

  cidr_ipv6   = "2001:db8:85a3::/64"
  from_port   = 80
  ip_protocol = "tcp"
  to_port     = 8080
}
`)
}

func testAccVPCSecurityGroupIngressRuleConfig_cidrIPv6Updated(rName string) string {
	return acctest.ConfigCompose(testAccVPCSecurityGroupRuleConfig_base(rName), `
resource "aws_vpc_security_group_ingress_rule" "test" {
  security_group_id = aws_security_group.test.id

  cidr_ipv6   = "2001:db8:85a3:2::/64"
  ip_protocol = "icmpv6"
}
`)
}

func testAccVPCSecurityGroupIngressRuleConfig_description(rName, description string) string {
	return acctest.ConfigCompose(testAccVPCSecurityGroupRuleConfig_base(rName), fmt.Sprintf(`
resource "aws_vpc_security_group_ingress_rule" "test" {
  security_group_id = aws_security_group.test.id

  cidr_ipv4   = "10.0.0.0/8"
  from_port   = 80
  ip_protocol = "tcp"
  to_port     = 8080

  description = %[1]q
}
`, description))
}

func testAccVPCSecurityGroupIngressRuleConfig_prefixListIDBase(rName string) string {
	return acctest.ConfigCompose(testAccVPCSecurityGroupRuleConfig_base(rName), fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test1" {
  vpc_id       = aws_vpc.test.id
  service_name = "com.amazonaws.${data.aws_region.current.name}.s3"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test2" {
  vpc_id       = aws_vpc.test.id
  service_name = "com.amazonaws.${data.aws_region.current.name}.dynamodb"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCSecurityGroupIngressRuleConfig_prefixListID(rName string) string {
	return acctest.ConfigCompose(testAccVPCSecurityGroupIngressRuleConfig_prefixListIDBase(rName), `
resource "aws_vpc_security_group_ingress_rule" "test" {
  security_group_id = aws_security_group.test.id

  prefix_list_id = aws_vpc_endpoint.test1.prefix_list_id
  from_port      = 80
  ip_protocol    = "tcp"
  to_port        = 8080
}
`)
}

func testAccVPCSecurityGroupIngressRuleConfig_prefixListIDUpdated(rName string) string {
	return acctest.ConfigCompose(testAccVPCSecurityGroupIngressRuleConfig_prefixListIDBase(rName), `
resource "aws_vpc_security_group_ingress_rule" "test" {
  security_group_id = aws_security_group.test.id

  prefix_list_id = aws_vpc_endpoint.test2.prefix_list_id
  from_port      = 80
  ip_protocol    = "tcp"
  to_port        = 8080
}
`)
}

func testAccVPCSecurityGroupIngressRuleConfig_referencedSecurityGroupID(rName string) string {
	return acctest.ConfigCompose(testAccVPCSecurityGroupRuleConfig_base(rName), `
resource "aws_vpc_security_group_ingress_rule" "test" {
  security_group_id = aws_security_group.test.id

  referenced_security_group_id = aws_security_group.test.id
  from_port                    = 80
  ip_protocol                  = "tcp"
  to_port                      = 8080
}
`)
}

func testAccVPCSecurityGroupIngressRuleConfig_referencedSecurityGroupIDUpdated(rName string) string {
	return acctest.ConfigCompose(testAccVPCSecurityGroupRuleConfig_base(rName), fmt.Sprintf(`
resource "aws_security_group" "test1" {
  vpc_id = aws_vpc.test.id
  name   = "%[1]s-1"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_security_group_ingress_rule" "test" {
  security_group_id = aws_security_group.test.id

  referenced_security_group_id = aws_security_group.test1.id
  from_port                    = 80
  ip_protocol                  = "tcp"
  to_port                      = 8080
}
`, rName))
}

func testAccVPCSecurityGroupIngressRuleConfig_referencedSecurityGroupIDPeerVPC(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), testAccVPCSecurityGroupRuleConfig_base(rName), fmt.Sprintf(`
resource "aws_vpc" "peer" {
  provider = "awsalternate"

  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "peer" {
  provider = "awsalternate"

  vpc_id = aws_vpc.peer.id
  name   = %[1]q

  tags = {
    Name = %[1]q
  }
}

data "aws_caller_identity" "peer" {
  provider = "awsalternate"
}

# Requester's side of the connection.
resource "aws_vpc_peering_connection" "test" {
  vpc_id        = aws_vpc.test.id
  peer_vpc_id   = aws_vpc.peer.id
  peer_owner_id = data.aws_caller_identity.peer.account_id
  peer_region   = %[2]q
  auto_accept   = false

  tags = {
    Name = %[1]q
  }
}

# Accepter's side of the connection.
resource "aws_vpc_peering_connection_accepter" "peer" {
  provider = "awsalternate"

  vpc_peering_connection_id = aws_vpc_peering_connection.test.id
  auto_accept               = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_security_group_ingress_rule" "test" {
  security_group_id = aws_security_group.test.id

  referenced_security_group_id = "${data.aws_caller_identity.peer.account_id}/${aws_security_group.peer.id}"
  from_port                    = 80
  ip_protocol                  = "tcp"
  to_port                      = 8080

  depends_on = [aws_vpc_peering_connection_accepter.peer]
}
`, rName, acctest.Region()))
}
