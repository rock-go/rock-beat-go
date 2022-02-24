package dns

import (
	"github.com/miekg/dns"
	"github.com/rock-go/rock/auxlib"
	"github.com/rock-go/rock/buffer"
	"github.com/rock-go/rock/json"
	"github.com/rock-go/rock/lua"
	"github.com/rock-go/rock/node"
	"github.com/rock-go/rock/region"
	"net"
)

type Tx struct {
	code   string
	name   string
	host   string
	src    uint16
	dst    uint16
	addr   net.Addr
	msg    dns.Msg
	buf    *buffer.Byte
	region *region.Info
}

func (tx *Tx) ToLValue() lua.LValue {
	return lua.NewAnyData(tx)
}

func (tx *Tx) Remote() string {
	return tx.addr.(*net.IPAddr).IP.String()
}

func (tx *Tx) QS2S(enc *json.Encoder, qq []dns.Question) {
	n := len(qq)
	if n == 0 {
		return
	}

	for i := 0; i < n; i++ {
		enc.Tab("")
		q := tx.msg.Question[i]
		enc.KV("name", q.Name)
		enc.KV("type", dns.TypeToString[q.Qtype])
		enc.KV("class", dns.ClassToString[q.Qclass])
		enc.End("},")
	}
}

func (tx *Tx) RR2S(enc *json.Encoder, r dns.RR) {
	switch v := r.(type) {
	case *dns.A:
		enc.KV("A", v.A.String())
	case *dns.AAAA:
		enc.KV("AAAA", v.AAAA.String())
	case *dns.CNAME:
		enc.KV("CNAME", v.Target)
	case *dns.NULL:
		enc.KV("DATA", v.Data)
	case *dns.HINFO:
		enc.KV("CPU", v.Cpu)
		enc.KV("OS", v.Os)

	case *dns.MB:
		enc.KV("MB", v.Mb)
	case *dns.MG:
		enc.KV("MB", v.Mg)
	case *dns.MINFO:
		enc.KV("RMAIL", v.Rmail)
		enc.KV("EMAIL", v.Email)
	case *dns.MR:
		enc.KV("MR", v.Mr)
	case *dns.MF:
		enc.KV("MF", v.Mf)
	case *dns.MD:
		enc.KV("MD", v.Md)
	case *dns.MX:
		enc.KV("preference", v.Preference)
		enc.KV("mx", v.Mx)

	case *dns.AFSDB:
		enc.KV("hostname", v.Hostname)
		enc.KV("sub", v.Subtype)

	case *dns.X25:
		enc.KV("x25", v.PSDNAddress)
	case *dns.RT:
		enc.KV("host", v.Host)
		enc.KV("preference", v.Preference)

	case *dns.NS:
		enc.KV("ns", v.Ns)

	case *dns.PTR:
		enc.KV("ptr", v.Ptr)
	case *dns.RP:
		enc.KV("mbox", v.Mbox)
		enc.KV("txt", v.Txt)

	case *dns.SOA:
		enc.KV("ns", v.Ns)
		enc.KV("mbox", v.Mbox)
		enc.KV("serial", v.Serial)
		enc.KV("refresh", v.Refresh)
		enc.KV("retry", v.Retry)
		enc.KV("expire", v.Expire)
		enc.KV("min_ttl", v.Minttl)

	case *dns.TXT:
		enc.KV("txt", v.Txt)
	case *dns.SPF:
		enc.KV("txt", v.Txt)
	case *dns.AVC:
		enc.KV("txt", v.Txt)
	case *dns.SRV:
		enc.KV("target", v.Target)
		enc.KV("priority", v.Priority)
		enc.KV("weight", v.Weight)
		enc.KV("port", v.Port)

	case *dns.NAPTR:
		enc.KV("order", v.Order)
		enc.KV("preference", v.Preference)
		enc.KV("flags", v.Flags)
		enc.KV("service", v.Service)
		enc.KV("regexp", v.Regexp)
		enc.KV("replace", v.Replacement)

	case *dns.CERT:
		enc.KV("type", dns.CertTypeToString[v.Type])
		enc.KV("tag", v.KeyTag)
		enc.KV("algorithm", v.Algorithm)
		enc.KV("certificate", v.Certificate)

	case *dns.DNAME:
		enc.KV("target", v.Target)
	case *dns.PX:
		enc.KV("preference", v.Preference)
		enc.KV("map822", v.Map822)
		enc.KV("mapx400", v.Mapx400)

	case *dns.GPOS:
		enc.KV("longitude", v.Longitude)
		enc.KV("latitude", v.Latitude)
		enc.KV("altitude", v.Altitude)

	case *dns.LOC:
		enc.KV("version", v.Version)
		enc.KV("size", v.Size)
		enc.KV("horiz_pre", v.HorizPre)
		enc.KV("longitude", v.Longitude)
		enc.KV("latitude", v.Latitude)
		enc.KV("altitude", v.Altitude)

	case *dns.SIG:
		enc.KV("type", dns.Type(v.TypeCovered).String())
		enc.KV("algorithm", v.Algorithm)
		enc.KV("labels", v.Labels)
		enc.KV("ttl", v.OrigTtl)
		enc.KV("expiration", v.Expiration)
		enc.KV("inception", v.Inception)
		enc.KV("tag", v.KeyTag)
		enc.KV("signer", v.SignerName)
		enc.KV("signature", v.Signature)

	case *dns.RRSIG:
		enc.KV("type", dns.Type(v.TypeCovered).String())
		enc.KV("algorithm", v.Algorithm)
		enc.KV("labels", v.Labels)
		enc.KV("ttl", v.OrigTtl)
		enc.KV("expiration", v.Expiration)
		enc.KV("inception", v.Inception)
		enc.KV("tag", v.KeyTag)
		enc.KV("signer", v.SignerName)
		enc.KV("signature", v.Signature)

	case *dns.NSEC:
		enc.KV("next", v.NextDomain)
		var ns []string
		for _, bit := range v.TypeBitMap {
			ns = append(ns, dns.Type(bit).String())
		}
		enc.Join("nsec", ns)

	case *dns.DLV:
		enc.KV("tag", v.KeyTag)
		enc.KV("algorithm", v.Algorithm)
		enc.KV("digestType", v.DigestType)
		enc.KV("digest", v.Digest)

	case *dns.CDS:
		enc.KV("tag", v.KeyTag)
		enc.KV("algorithm", v.Algorithm)
		enc.KV("digestType", v.DigestType)
		enc.KV("digest", v.Digest)

	case *dns.DS:
		enc.KV("tag", v.KeyTag)
		enc.KV("algorithm", v.Algorithm)
		enc.KV("digestType", v.DigestType)
		enc.KV("digest", v.Digest)

	case *dns.KX:
		enc.KV("preference", v.Preference)
		enc.KV("exchanger", v.Exchanger)

	case *dns.TA:
		enc.KV("tag", v.KeyTag)
		enc.KV("algorithm", v.Algorithm)
		enc.KV("digestType", v.DigestType)
		enc.KV("digest", v.Digest)

	case *dns.TALINK:
		enc.KV("prev", v.PreviousName)
		enc.KV("next", v.NextName)

	case *dns.SSHFP:
		enc.KV("algorithm", v.Algorithm)
		enc.KV("type", v.Type)
		enc.KV("hex", v.FingerPrint)

	case *dns.KEY:
		enc.KV("flags", v.Flags)
		enc.KV("protocol", v.Protocol)
		enc.KV("algorithm", v.Algorithm)
		enc.KV("publicKey", v.PublicKey)

	case *dns.CDNSKEY:
		enc.KV("flags", v.Flags)
		enc.KV("protocol", v.Protocol)
		enc.KV("algorithm", v.Algorithm)
		enc.KV("publicKey", v.PublicKey)

	case *dns.DNSKEY:
		enc.KV("flags", v.Flags)
		enc.KV("protocol", v.Protocol)
		enc.KV("algorithm", v.Algorithm)
		enc.KV("publicKey", v.PublicKey)

	case *dns.RKEY:
		enc.KV("flags", v.Flags)
		enc.KV("protocol", v.Protocol)
		enc.KV("algorithm", v.Algorithm)
		enc.KV("publicKey", v.PublicKey)

	case *dns.NSAPPTR:
		enc.KV("ptr", v.Ptr)

	case *dns.NSEC3:
		enc.KV("hash", v.Hash)
		enc.KV("flags", v.Flags)
		enc.KV("Iterations", v.Iterations)
		enc.KV("saltLength", v.SaltLength)
		enc.KV("salt", v.Salt)
		enc.KV("hashLength", v.HashLength)
		enc.KV("nextDomain", v.NextDomain)

		var ns []string
		for _, t := range v.TypeBitMap {
			ns = append(ns, dns.Type(t).String())
		}
		enc.Join("nsec", ns)

	case *dns.NSEC3PARAM:
		enc.KV("hash", v.Hash)
		enc.KV("flags", v.Flags)
		enc.KV("Iterations", v.Iterations)
		enc.KV("saltLength", v.SaltLength)
		enc.KV("salt", v.Salt)

	case *dns.TKEY:
		enc.KV("algorithm", v.Algorithm)
		enc.KV("inception", v.Inception)
		enc.KV("expiration", v.Expiration)
		enc.KV("mode", v.Mode)
		enc.KV("error", v.Error)
		enc.KV("size", v.KeySize)
		enc.KV("key", v.Key)
		enc.KV("length", v.OtherLen)
		enc.KV("data", v.OtherData)

	case *dns.RFC3597:
		enc.KV("hash", v.Rdata)

	case *dns.URI:
		enc.KV("digest", v.Target)
		enc.KV("priority", v.Priority)
		enc.KV("weight", v.Weight)

	case *dns.DHCID:
		enc.KV("digest", v.Digest)

	case *dns.TLSA:
		enc.KV("usage", v.Usage)
		enc.KV("selector", v.Selector)
		enc.KV("match_type", v.MatchingType)
		enc.KV("cert", v.Certificate)

	case *dns.SMIMEA:
		enc.KV("usage", v.Usage)
		enc.KV("selector", v.Selector)
		enc.KV("match_type", v.MatchingType)
		enc.KV("cert", v.Certificate)

	case *dns.HIP:
		enc.KV("pub_Alg", v.PublicKeyAlgorithm)
		enc.KV("pub_length", v.PublicKeyLength)
		enc.KV("hit", v.Hit)
		enc.KV("pub", v.PublicKey)
		enc.Join("servers", v.RendezvousServers)

	case *dns.NINFO:
		enc.Join("zs", v.ZSData)

	case *dns.NID:
		enc.KV("preference", v.Preference)
		enc.KV("node", v.NodeID)

	case *dns.L32:
		enc.KV("preference", v.Preference)
		enc.KV("locator32", v.Locator32)

	case *dns.L64:
		enc.KV("preference", v.Preference)
		enc.KV("locator64", v.Locator64)

	case *dns.LP:
		enc.KV("preference", v.Preference)
		enc.KV("rqdn", v.Fqdn)

	case *dns.EUI48:
		enc.KV("address", v.Address)

	case *dns.CAA:
		enc.KV("flag", v.Flag)
		enc.KV("tag", v.Tag)
		enc.KV("value", v.Value)

	case *dns.UID:
		enc.KV("uid", v.Uid)

	case *dns.GID:
		enc.KV("gid", v.Gid)

	case *dns.UINFO:
		enc.KV("uinfo", v.Uinfo)

	case *dns.NIMLOC:
		enc.KV("locator", v.Locator)

	case *dns.OPENPGPKEY:
		enc.KV("public", v.PublicKey)

	case *dns.CSYNC:
		enc.KV("serial", v.Serial)
		enc.KV("flags", v.Flags)
		var ns []string
		for _, t := range v.TypeBitMap {
			ns = append(ns, dns.Type(t).String())
		}
		enc.Join("nsec", ns)

	case *dns.ZONEMD:
		enc.KV("serial", v.Serial)
		enc.KV("scheme", v.Scheme)
		enc.KV("hash", v.Hash)
		enc.KV("digest", v.Digest)

	case *dns.APL:
		enc.Arr("apl")
		for _, p := range v.Prefixes {
			enc.Tab("")
			enc.KV("negation", p.Negation)
			enc.KV("network", p.Network)
			enc.End("},")
		}
		enc.End("],")

	default:

	}

}

func (tx *Tx) RS2S(enc *json.Encoder, rr []dns.RR) {
	n := len(rr)
	if n == 0 {
		return
	}

	for i := 0; i < n; i++ {
		r := rr[i]
		enc.Tab("")
		h := r.Header()
		enc.KV("name", h.Name)
		enc.KV("type", dns.Type(h.Rrtype).String())
		enc.KV("class", dns.Class(h.Class).String())
		enc.KV("ttl", h.Ttl)
		enc.KV("length", h.Rdlength)
		tx.RR2S(enc, r)
		enc.End("},")
	}
}

func (tx *Tx) String() string {
	enc := json.NewEncoder()

	enc.Tab("")
	enc.KV("ID", node.ID())
	enc.KV("inet", node.LoadAddr())
	enc.KV("remote", tx.Remote())
	enc.KV("region", tx.region.Byte())
	enc.KV("host", tx.host)

	enc.KV("dns_id", tx.msg.Id)
	enc.KV("response", tx.msg.Response)
	enc.KV("op_code", tx.msg.Opcode)
	enc.KV("authoritative", tx.msg.Authoritative)
	enc.KV("truncated", tx.msg.Truncated)
	enc.KV("recursionDesired", tx.msg.RecursionDesired)
	enc.KV("recursionDesired", tx.msg.RecursionAvailable)
	enc.KV("zero", tx.msg.Zero)
	enc.KV("authenticated", tx.msg.AuthenticatedData)
	enc.KV("disable", tx.msg.CheckingDisabled)
	enc.KV("r_code", tx.msg.Rcode)
	enc.KV("compress", tx.msg.Compress)

	enc.Arr("question")
	tx.QS2S(enc, tx.msg.Question)
	enc.End("],")

	enc.Arr("answer")
	tx.RS2S(enc, tx.msg.Answer)
	enc.End("],")

	enc.Arr("extra")
	tx.RS2S(enc, tx.msg.Extra)
	enc.End("],")

	enc.Arr("ns")
	tx.RS2S(enc, tx.msg.Ns)
	enc.End("],")

	enc.End("}")
	tx.buf = enc.Buffer()
	return auxlib.B2S(enc.Bytes())
}

func (tx *Tx) Index(L *lua.LState, key string) lua.LValue {
	return lua.LNil
}
